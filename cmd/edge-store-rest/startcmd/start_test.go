/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"net/http"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type mockServer struct{}

func (s *mockServer) ListenAndServe(host string, handler http.Handler) error {
	return nil
}

func TestStartCmdContents(t *testing.T) {
	startCmd := GetStartCmd(&mockServer{})

	require.Equal(t, "start", startCmd.Use)
	require.Equal(t, "Start edge-store", startCmd.Short)
	require.Equal(t, "Start edge-store", startCmd.Long)

	checkFlagPropertiesCorrect(t, startCmd, hostURLFlagName, hostURLFlagShorthand, hostURLFlagUsage)
	checkFlagPropertiesCorrect(t, startCmd, databaseTypeFlagName, databaseTypeFlagShorthand, databaseTypeFlagUsage)
	checkFlagPropertiesCorrect(t, startCmd, databaseURLFlagName, databaseURLFlagShorthand, databaseURLFlagUsage)
}

func TestStartCmdWithBlankHostArg(t *testing.T) {
	startCmd := GetStartCmd(&mockServer{})

	args := []string{"--" + hostURLFlagName, "", "--" + databaseTypeFlagName, "memstore"}
	startCmd.SetArgs(args)

	err := startCmd.Execute()

	require.Equal(t, errMissingHostURL.Error(), err.Error())
}

func TestStartCmdWithMissingHostArg(t *testing.T) {
	startCmd := GetStartCmd(&mockServer{})

	err := startCmd.Execute()

	require.Equal(t,
		"Neither host-url (command line flag) nor EDGE-STORE_HOST_URL (environment variable) have been set.",
		err.Error())
}

func TestStartEdgeStoreWithBlankHost(t *testing.T) {
	parameters := &edgeStoreParameters{hostURL: ""}

	err := startEdgeStore(parameters)
	require.NotNil(t, err)
	require.Equal(t, errMissingHostURL, err)
}

func TestStartCmdValidArgs(t *testing.T) {
	startCmd := GetStartCmd(&mockServer{})

	args := []string{"--" + hostURLFlagName, "localhost:8080", "--" + databaseTypeFlagName, "memstore"}
	startCmd.SetArgs(args)

	err := startCmd.Execute()

	require.Nil(t, err)
}

func TestStartCmdValidArgsEnvVar(t *testing.T) {
	startCmd := GetStartCmd(&mockServer{})

	err := os.Setenv(hostURLEnvKey, "localhost:8080")
	require.Nil(t, err)

	err = os.Setenv(databaseTypeEnvKey, "memstore")
	require.Nil(t, err)

	err = startCmd.Execute()

	require.Nil(t, err)
}

func checkFlagPropertiesCorrect(t *testing.T, cmd *cobra.Command, flagName, flagShorthand, flagUsage string) {
	flag := cmd.Flag(flagName)

	require.NotNil(t, flag)
	require.Equal(t, flagName, flag.Name)
	require.Equal(t, flagShorthand, flag.Shorthand)
	require.Equal(t, flagUsage, flag.Usage)
}
