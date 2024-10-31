// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklogtest"
)

func TestServerGetProviderSchema(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.GetProviderSchemaRequest
		expectedError    error
		expectedResponse *tfprotov5.GetProviderSchemaResponse
	}{
		"datasourceschemas": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
											resp.Schema = datasourceschema.Schema{
												Attributes: map[string]datasourceschema.Attribute{
													"test1": datasourceschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source1"
										},
									}
								},
								func() datasource.DataSource {
									return &testprovider.DataSource{
										SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
											resp.Schema = datasourceschema.Schema{
												Attributes: map[string]datasourceschema.Attribute{
													"test2": datasourceschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov5.Schema{
					"test_data_source1": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_data_source2": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Functions:                map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasourceschemas-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
											resp.Schema = datasourceschema.Schema{
												Attributes: map[string]datasourceschema.Attribute{
													"test1": datasourceschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source"
										},
									}
								},
								func() datasource.DataSource {
									return &testprovider.DataSource{
										SchemaMethod: func(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
											resp.Schema = datasourceschema.Schema{
												Attributes: map[string]datasourceschema.Attribute{
													"test2": datasourceschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = "test_data_source"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Data Source Type Defined",
						Detail: "The test_data_source data source type name was returned for multiple data sources. " +
							"Data source type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"datasourceschemas-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						DataSourcesMethod: func(_ context.Context) []func() datasource.DataSource {
							return []func() datasource.DataSource{
								func() datasource.DataSource {
									return &testprovider.DataSource{
										MetadataMethod: func(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Data Source Type Name Missing",
						Detail: "The *testprovider.DataSource DataSource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralschemas": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = ephemeralschema.Schema{
												Attributes: map[string]ephemeralschema.Attribute{
													"test1": ephemeralschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource1"
										},
									}
								},
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = ephemeralschema.Schema{
												Attributes: map[string]ephemeralschema.Attribute{
													"test2": ephemeralschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{
					"test_ephemeral_resource1": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_ephemeral_resource2": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralschemas-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = ephemeralschema.Schema{
												Attributes: map[string]ephemeralschema.Attribute{
													"test1": ephemeralschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
									}
								},
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										SchemaMethod: func(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
											resp.Schema = ephemeralschema.Schema{
												Attributes: map[string]ephemeralschema.Attribute{
													"test2": ephemeralschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = "test_ephemeral_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Ephemeral Resource Type Defined",
						Detail: "The test_ephemeral_resource ephemeral resource type name was returned for multiple ephemeral resources. " +
							"Ephemeral resource type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"ephemeralschemas-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						EphemeralResourcesMethod: func(_ context.Context) []func() ephemeral.EphemeralResource {
							return []func() ephemeral.EphemeralResource{
								func() ephemeral.EphemeralResource {
									return &testprovider.EphemeralResource{
										MetadataMethod: func(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Ephemeral Resource Type Name Missing",
						Detail: "The *testprovider.EphemeralResource EphemeralResource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functions": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "function1"
										},
									}
								},
								func() function.Function {
									return &testprovider.Function{
										DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "function2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Functions: map[string]*tfprotov5.Function{
					"function1": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
							Type: tftypes.String,
						},
					},
					"function2": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functions-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction" // intentionally duplicate
										},
									}
								},
								func() function.Function {
									return &testprovider.Function{
										DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction" // intentionally duplicate
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Function Name Defined",
						Detail: "The testfunction function name was returned for multiple functions. " +
							"Function names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"functions-empty-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "" // intentionally empty
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Function Name Missing",
						Detail: "The *testprovider.Function Function returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"provider": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						SchemaMethod: func(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
							resp.Schema = providerschema.Schema{
								Attributes: map[string]providerschema.Attribute{
									"test": providerschema.StringAttribute{
										Required: true,
									},
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Functions:                map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
							{
								Name:     "test",
								Required: true,
								Type:     tftypes.String,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"providermeta": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithMetaSchema{
						Provider: &testprovider.Provider{},
						MetaSchemaMethod: func(_ context.Context, _ provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
							resp.Schema = metaschema.Schema{
								Attributes: map[string]metaschema.Attribute{
									"test": metaschema.StringAttribute{
										Required: true,
									},
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Functions:                map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ProviderMeta: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
							{
								Name:     "test",
								Required: true,
								Type:     tftypes.String,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resourceschemas": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = resourceschema.Schema{
												Attributes: map[string]resourceschema.Attribute{
													"test1": resourceschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource1"
										},
									}
								},
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = resourceschema.Schema{
												Attributes: map[string]resourceschema.Attribute{
													"test2": resourceschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Functions:                map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{
					"test_resource1": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test1",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_resource2": {
						Block: &tfprotov5.SchemaBlock{
							Attributes: []*tfprotov5.SchemaAttribute{
								{
									Name:     "test2",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resourceschemas-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = resourceschema.Schema{
												Attributes: map[string]resourceschema.Attribute{
													"test1": resourceschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
								func() resource.Resource {
									return &testprovider.Resource{
										SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
											resp.Schema = resourceschema.Schema{
												Attributes: map[string]resourceschema.Attribute{
													"test2": resourceschema.StringAttribute{
														Required: true,
													},
												},
											}
										},
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = "test_resource"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Resource Type Defined",
						Detail: "The test_resource resource type name was returned for multiple resources. " +
							"Resource type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
		"resourceschemas-empty-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ResourcesMethod: func(_ context.Context) []func() resource.Resource {
							return []func() resource.Resource{
								func() resource.Resource {
									return &testprovider.Resource{
										MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
											resp.TypeName = ""
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetProviderSchemaRequest{},
			expectedResponse: &tfprotov5.GetProviderSchemaResponse{
				DataSourceSchemas:        map[string]*tfprotov5.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov5.Schema{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Resource Type Name Missing",
						Detail: "The *testprovider.Resource Resource returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
				Provider: &tfprotov5.Schema{
					Block: &tfprotov5.SchemaBlock{},
				},
				ResourceSchemas: map[string]*tfprotov5.Schema{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					MoveResourceState:         true,
					PlanDestroy:               true,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.GetProviderSchema(context.Background(), new(tfprotov5.GetProviderSchemaRequest))

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}

func TestServerGetProviderSchema_logging(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)
	ctx = logging.InitContext(ctx)

	testServer := &Server{
		FrameworkServer: fwserver.Server{
			Provider: &testprovider.Provider{},
		},
	}

	_, err := testServer.GetProviderSchema(ctx, new(tfprotov5.GetProviderSchemaRequest))

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	entries, err := tfsdklogtest.MultilineJSONDecode(&output)

	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	expectedEntries := []map[string]interface{}{
		{
			"@level":   "trace",
			"@message": "Checking ProviderSchema lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider Schema",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider Schema",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ResourceTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ProviderTypeName lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider Metadata",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider Metadata",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider Resources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider Resources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking DataSourceTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ProviderTypeName lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider Metadata",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider Metadata",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider DataSources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider DataSources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking FunctionTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking EphemeralResourceFuncs lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ProviderTypeName lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider Metadata",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider Metadata",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Calling provider defined Provider EphemeralResources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Called provider defined Provider EphemeralResources",
			"@module":  "sdk.framework",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}
