// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/log"
	"github.com/googleapis/genai-toolbox/internal/server"
)

// TestServerOptions holds a configured server and resources for testing
type TestServerOptions struct {
	Server *server.Server
	ctx    context.Context
	cancel context.CancelFunc
	
	// serverAddr is the address:port where the server is listening
	serverAddr string
}

// Close cleans up resources allocated for the test server
func (opts *TestServerOptions) Close() {
	if opts.cancel != nil {
		opts.cancel()
	}
}

// RunIntegrationTests determines if integration tests should be run based on
// the RUN_INTEGRATION_TESTS environment variable.
func RunIntegrationTests() bool {
	return os.Getenv("RUN_INTEGRATION_TESTS") == "1"
}

// NewTestServerOptions creates a new server with provided source and tool configs.
// It returns a struct containing the server and a cleanup function.
func NewTestServerOptions(sourceYaml, toolYaml string) *TestServerOptions {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Combine source and tool YAML configurations
	configYaml := sourceYaml + "\n" + toolYaml
	
	// Create a struct to hold the parsed configuration
	type toolsFile struct {
		Sources      server.SourceConfigs      `yaml:"sources"`
		AuthServices server.AuthServiceConfigs `yaml:"authServices"`
		Tools        server.ToolConfigs        `yaml:"tools"`
		Toolsets     server.ToolsetConfigs     `yaml:"toolsets"`
	}
	
	// Debug print the YAML before parsing
	fmt.Printf("Config YAML: %s\n", configYaml)
	
	var tf toolsFile
	err := yaml.UnmarshalContext(ctx, []byte(configYaml), &tf, yaml.Strict())
	if err != nil {
		panic(fmt.Sprintf("Failed to parse config: %v", err))
	}
	
	// Create server configuration
	cfg := server.ServerConfig{
		Address:            "localhost",
		Port:               0, // Use a random port
		SourceConfigs:      tf.Sources,
		AuthServiceConfigs: tf.AuthServices,
		ToolConfigs:        tf.Tools,
		ToolsetConfigs:     tf.Toolsets,
	}
	
	// Create logger
	logger, err := log.NewStdLogger(os.Stdout, os.Stderr, "DEBUG")
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	
	// Create a new server with the parsed configuration
	srv, err := server.NewServer(ctx, cfg, logger)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test server: %v", err))
	}
	
	// Start the server but don't block
	if err := srv.Listen(ctx); err != nil {
		cancel()
		panic(fmt.Sprintf("Failed to start listener: %v", err))
	}

	// Get the server address
	addr := GetServerListenerAddr(srv)
	if addr == nil {
		cancel()
		panic("Failed to get server listener address")
	}
	
	// Start serving in a goroutine
	go func() {
		if err := srv.Serve(ctx); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()
	
	return &TestServerOptions{
		Server:     srv,
		ctx:        ctx,
		cancel:     cancel,
		serverAddr: addr.String(),
	}
}

// CallTool invokes a tool via the server API and returns the result
func CallTool(srv *server.Server, toolName string, params map[string]interface{}) (interface{}, error) {
	// Create the JSON payload
	payload := map[string]interface{}{
		"params": params,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}
	
	// Get server address
	addr := GetServerListenerAddr(srv)
	if addr == nil {
		return nil, fmt.Errorf("server listener not available")
	}
	
	// Build the URL
	url := fmt.Sprintf("http://%s/api/tools/%s", addr.String(), toolName)
	
	// Create and execute the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	
	// Check for error status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status %d: %s", resp.StatusCode, body)
	}
	
	// Parse the response JSON
	var result struct {
		Result interface{} `json:"result"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}
	
	return result.Result, nil
}
