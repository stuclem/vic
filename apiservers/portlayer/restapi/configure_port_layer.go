// Copyright 2016 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package restapi

import (
	//"io"
	"net/http"

	errors "github.com/go-swagger/go-swagger/errors"
	httpkit "github.com/go-swagger/go-swagger/httpkit"

	"github.com/vmware/vic/apiservers/portlayer/restapi/handlers"
	"github.com/vmware/vic/apiservers/portlayer/restapi/operations"
)

// This file is safe to edit. Once it exists it will not be overwritten

type portlayerhandlers struct {
	storageHandlers handlers.StorageHandlersImpl
	miscHandlers    handlers.MiscHandlersImpl
}

func configureAPI(api *operations.PortLayerAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	api.BinConsumer = httpkit.ByteStreamConsumer()

	api.JSONConsumer = httpkit.JSONConsumer()

	api.JSONProducer = httpkit.JSONProducer()

	allhandlers := portlayerhandlers{}

	allhandlers.storageHandlers.Configure(api)
	allhandlers.miscHandlers.Configure(api)

	api.ServerShutdown = func() {}
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}