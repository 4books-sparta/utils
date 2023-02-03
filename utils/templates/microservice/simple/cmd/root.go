package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Name    = "--module---microservice"
	Version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:   "sparta_ms_--service-name--",
	Short: "Mother command",
}

func init() {
	rootCmd.Flags().Uint16("port", 80, "bind port")
	rootCmd.PersistentFlags().String("db_vendor", "postgres", "database vendor")
	rootCmd.PersistentFlags().String("db_host", "localhost", "database host")
	rootCmd.PersistentFlags().Int("db_port", 5432, "database port")
	rootCmd.PersistentFlags().String("db_user", "-", "database username")
	rootCmd.PersistentFlags().String("db_password", "-", "database password")
	rootCmd.PersistentFlags().String("db_name", "-", "database name")
	rootCmd.PersistentFlags().String("db_ssl", "disable", "database ssl mode")
	rootCmd.PersistentFlags().Int("db_timeout", 2, "database timeout in minutes")
	rootCmd.PersistentFlags().Int("db_idle", 10, "database max idle connections")
	rootCmd.PersistentFlags().Int("db_open", 80, "database max open connections")

	_ = viper.BindPFlags(rootCmd.PersistentFlags())
	viper.AutomaticEnv()
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
