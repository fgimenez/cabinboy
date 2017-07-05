package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	helmImage = "quay.io/giantswarm/docker-helm:006b0db51ec484be8b1bd49990804784a9737ece"
)

var (
	// promoteCmd represents the promote command
	promoteCmd = &cobra.Command{
		Use:   "promote",
		Short: "Move chart among channels",
		Long: `This command let's you change the channel of a chart.

By default it just needs to know the chart name, the origin channel and chart version`,
		Run: runPromote,
	}
	toChannel    string
	username     string
	password     string
	project      string
	organisation string
	version      string
)

func init() {
	RootCmd.AddCommand(promoteCmd)

	promoteCmd.Flags().StringVar(&toChannel, "to", "", "Target channel")

	promoteCmd.Flags().StringVar(&username, "username", os.Getenv("QUAY_USERNAME"), "username to use to login to docker registry")
	promoteCmd.Flags().StringVar(&password, "password", os.Getenv("QUAY_PASSWORD"), "password to use to login to docker registry")
	promoteCmd.Flags().StringVar(&project, "project", os.Getenv("CIRCLE_PROJECT_REPONAME")+"-chart", "Project name")
	promoteCmd.Flags().StringVar(&organisation, "organisation", os.Getenv("CIRCLE_PROJECT_USERNAME"), "Organisation")
	promoteCmd.Flags().StringVar(&version, "version", "", "Chart version")
}

func runPromote(cmd *cobra.Command, args []string) {
	log.Println("promote called")
	if err := helmLogin(); err != nil {
		log.Fatalf("error login to helm: %v", err)
	}

	if err := helmChannelPromotion(); err != nil {
		log.Fatalf("error changing channels in helm: %v", err)
	}
}

func helmLogin() error {
	cnrDir, err := cnrDirectory()
	if err != nil {
		return err
	}
	log.Println("helm login")
	cmd := exec.Command("docker", "run",
		"-v", cnrDir+":/root/.cnr",
		helmImage,
		"registry",
		"login",
		"--user="+username,
		"--password="+password,
		"quay.io")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func helmChannelPromotion() error {
	log.Println("helm channel")
	cnrDir, err := cnrDirectory()
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", "run",
		"-v", cnrDir+":/root/.cnr",
		helmImage,
		"registry",
		"channel",
		"--channel "+toChannel,
		"--set-release",
		fmt.Sprintf("quay.io/%s/%s@%s", organisation, project, version))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func cnrDirectory() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(user.HomeDir, ".cnr"), nil
}
