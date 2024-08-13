package cmd

import "github.com/spf13/cobra"

var SignupCmd = &cobra.Command{
	Use: "signup",
	Short: "User can signup via github",
	Run: runSignup,
}

func runSignup(cmd *cobra.Command, args []string){

}