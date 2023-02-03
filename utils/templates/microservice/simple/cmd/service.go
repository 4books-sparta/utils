package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/4books-sparta/utils"
	"github.com/4books-sparta/utils/requests"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"goji.io"
	"goji.io/pat"

	"--module--/pkg/service"
	"--module--/pkg/service/--service-name--"

)

var serviceCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Start the micro-service",
	Run:   runMicroservice,
}

func init() {
	serviceCmd.Flags().Uint16("port", 0, "bind port")

	serviceCmd.Flags().String("redis_host", "localhost", "redis host")
	serviceCmd.Flags().String("redis_cache_enabled", "no", "enable cache")
	serviceCmd.Flags().String("redis_cluster", "no", "enable cluster mode yes/no")
	serviceCmd.Flags().String("redis_auth_token", "", "password")

	serviceCmd.Flags().String("prom_user", "--prom-user--", "prometheus basic auth user")
	serviceCmd.Flags().String("prom_password", "--prom-pass--", "prometheus basic auth password")
	serviceCmd.Flags().String("prom_realm", "prometheus", "prometheus basic auth realm")

	serviceCmd.Flags().String("is_production", "n", "for logging purposes")

	_ = viper.BindPFlags(serviceCmd.Flags())
	viper.AutomaticEnv()

	rootCmd.AddCommand(serviceCmd)
}

func runMicroservice(_ *cobra.Command, _ []string) {
	port := viper.GetInt("port")
	logger := utils.NewLogger()

	// database configuration
	db, err := utils.NewDatabase(utils.GetDbConfig())
	if err != nil {
		_ = logger.Log("err", err)
		panic(err)
	}

	redisConfig := utils.GetRedisConfig()
	redisClient, noRed := utils.NewRedisClient(redisConfig)
	if noRed == nil {
		defer func() {
			_ = redisClient.Close()
		}()
	}

	var rep utils.ErrorReporter

	dsn := viper.GetString("sentry_dsn")
	if dsn == "" {
		rep = utils.NewLoggingReporter(logger)
	} else {
		rep, err = utils.NewSentryReporter(dsn)
		if err != nil {
			_ = logger.Log("err", err)
			panic(err)
		}
	}

	met := utils.PrometheusMetric(service.Name, service.Service)
	_ = --service-name--.New(--service-name--.NewSqlRepo(db), rep, met, redisClient) //Main service


	authZ := authorizer{}

	fw := utils.NewForwarder(authZ)
	mux := goji.NewMux()

	mux.Handle(pat.Get("/metrics"), utils.MakePrometheusHandler(
		viper.GetString("prom_user"), viper.GetString("prom_password"),
		viper.GetString("prom_realm"),
	))

	mux.HandleFunc(pat.Options("/*"), utils.Preflight)

	mux.Handle(pat.Get("/_probe"), fw.Forward(
		utils.MakeProbeEndpoint(), requests.DecodeEmptyRequest))


	fmt.Println("Sparta ", Name, Version, " is LISTENING on port ", port)
	_ = logger.Log(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))

}

type authorizer struct {
}

func (a authorizer) Authorize(_ context.Context, _ *http.Request) utils.Authorization {
	return utils.Authorization{}
}
