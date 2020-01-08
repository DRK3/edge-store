/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/trustbloc/edge-store/pkg/restapi/auth"
	cmdutils "github.com/trustbloc/edge-store/pkg/utils/cmd"
)

const (
	hostURLFlagName      = "host-url"
	hostURLFlagShorthand = "u"
	hostURLFlagUsage     = "URL to run the edge-auth instance on. Format: HostName:Port."
	hostURLEnvKey        = "EDGE-AUTH_HOST_URL"
)

var errMissingHostURL = errors.New("host URL not provided")

type server interface {
	ListenAndServe(host string, router http.Handler) error
}

// HTTPServer represents an actual HTTP server implementation.
type HTTPServer struct{}

// ListenAndServe starts the server using the standard Go HTTP server implementation.
func (s *HTTPServer) ListenAndServe(host string, router http.Handler) error {
	return http.ListenAndServe(host, router)
}

type edgeAuthParameters struct {
	srv     server
	hostURL string
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
		Short: "Start edge-auth",
		Long:  "Start edge-auth",
		RunE: func(cmd *cobra.Command, args []string) error {
			hostURL, err := cmdutils.GetUserSetVar(cmd, hostURLFlagName, hostURLEnvKey, false)
			if err != nil {
				return err
			}
			parameters := &edgeAuthParameters{
				srv:     srv,
				hostURL: hostURL,
			}
			return startEdgeAuth(parameters)
		},
	}
}

func createFlags(startCmd *cobra.Command) {
	startCmd.Flags().StringP(hostURLFlagName, hostURLFlagShorthand, "", hostURLFlagUsage)
}

func startEdgeAuth(parameters *edgeAuthParameters) error {
	if parameters.hostURL == "" {
		return errMissingHostURL
	}

	authService, err := auth.New()
	if err != nil {
		return err
	}

	handlers := authService.GetOperations()
	router := mux.NewRouter()

	for _, handler := range handlers {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	err = parameters.srv.ListenAndServe(parameters.hostURL, router)

	return err
}
