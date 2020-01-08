/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/trustbloc/edge-store/pkg/restapi/edv"
	"github.com/trustbloc/edge-store/pkg/restapi/edv/operation"
	"github.com/trustbloc/edge-store/pkg/storage"
	"github.com/trustbloc/edge-store/pkg/storage/couchdb"
	"github.com/trustbloc/edge-store/pkg/storage/mem"
	cmdutils "github.com/trustbloc/edge-store/pkg/utils/cmd"
)

const (
	hostURLFlagName      = "host-url"
	hostURLFlagShorthand = "o"
	hostURLFlagUsage     = "URL to run the edge-store instance on. Format: HostName:Port."
	hostURLEnvKey        = "EDGE-STORE_HOST_URL"

	databaseTypeFlagName      = "database-type"
	databaseTypeFlagShorthand = "t"
	databaseTypeFlagUsage     = "The type of database to use internally in the EDV. Supported options: memstore, couchdb"
	databaseTypeEnvKey        = "EDGE-STORE_DATABASE_TYPE"

	databaseURLFlagName      = "database-url"
	databaseURLFlagShorthand = "l"
	databaseURLFlagUsage     = "The URL of the database. Not needed if using memstore."
	databaseURLEnvKey        = "EDGE-STORE_DATABASE_URL"
)

var errMissingHostURL = errors.New("host URL not provided")

type edgeStoreParameters struct {
	srv          server
	hostURL      string
	databaseType string
	databaseURL  string
}

type server interface {
	ListenAndServe(host string, router http.Handler) error
}

// HTTPServer represents an actual HTTP server implementation.
type HTTPServer struct{}

// ListenAndServe starts the server using the standard Go HTTP server implementation.
func (s *HTTPServer) ListenAndServe(host string, router http.Handler) error {
	return http.ListenAndServe(host, router)
}

// GetStartCmd returns the Cobra start command.
func GetStartCmd(srv server) *cobra.Command {
	startCmd := createStartCmd(srv)

	createFlags(startCmd)

	return startCmd
}

func createStartCmd(srv server) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start edge-store",
		Long:  "Start edge-store",
		RunE: func(cmd *cobra.Command, args []string) error {
			hostURL, err := cmdutils.GetUserSetVar(cmd, hostURLFlagName, hostURLEnvKey, false)
			if err != nil {
				return err
			}
			databaseType, err := cmdutils.GetUserSetVar(cmd, databaseTypeFlagName, databaseTypeEnvKey, false)
			if err != nil {
				return err
			}
			databaseURL, err := cmdutils.GetUserSetVar(cmd, databaseURLFlagName, databaseURLEnvKey, true)
			if err != nil {
				return err
			}
			parameters := &edgeStoreParameters{
				srv:          srv,
				hostURL:      hostURL,
				databaseType: databaseType,
				databaseURL:  databaseURL,
			}
			return startEdgeStore(parameters)
		},
	}
}

func createFlags(startCmd *cobra.Command) {
	startCmd.Flags().StringP(hostURLFlagName, hostURLFlagShorthand, "", hostURLFlagUsage)
	startCmd.Flags().StringP(databaseTypeFlagName, databaseTypeFlagShorthand, "", databaseTypeFlagUsage)
	startCmd.Flags().StringP(databaseURLFlagName, databaseURLFlagShorthand, "", databaseURLFlagUsage)
}

func startEdgeStore(parameters *edgeStoreParameters) error {
	if parameters.hostURL == "" {
		return errMissingHostURL
	}

	var provider storage.Provider

	var handlers []operation.Handler

	if strings.EqualFold(parameters.databaseType, "memstore") {
		provider = mem.NewProvider()
	} else if parameters.databaseType == "couchdb" {
		couchDBProvider, err := couchdb.NewProvider(parameters.databaseURL)
		if err != nil {
			return err
		}
		provider = couchDBProvider
	}

	edvService, err := edv.New(provider)
	if err != nil {
		return err
	}

	handlers = edvService.GetOperations()

	router := mux.NewRouter()

	for _, handler := range handlers {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	err = parameters.srv.ListenAndServe(parameters.hostURL, router)

	return err
}
