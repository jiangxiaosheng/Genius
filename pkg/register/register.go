package register

import (
	genius "github.com/genius/pkg/schedule"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
)

// Register register to the sig-scheduler API.
func Register() *cobra.Command {
	return app.NewSchedulerCommand(
		app.WithPlugin(genius.SchedulerName, genius.New))
}
