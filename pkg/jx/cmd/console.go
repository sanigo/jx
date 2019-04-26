package cmd

import (
	"fmt"

	"github.com/jenkins-x/jx/pkg/kube"

	"github.com/spf13/cobra"

	"github.com/jenkins-x/jx/pkg/jx/cmd/opts"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/pkg/browser"
)

type ConsoleOptions struct {
	GetURLOptions

	OnlyViewURL     bool
	ClassicMode     bool
	JenkinsSelector opts.JenkinsSelectorOptions
}

const (
	BlueOceanPath = "/blue"
)

var (
	console_long = templates.LongDesc(`
		Opens the Jenkins X console in a browser.`)
	console_example = templates.Examples(`
		# Open the Jenkins X console in a browser
		jx console

		# Print the Jenkins X console URL but do not open a browser
		jx console -u
		
		# Open the Jenkins X console in a browser using the classic skin
		jx console --classic`)
)

func NewCmdConsole(commonOpts *opts.CommonOptions) *cobra.Command {
	options := &ConsoleOptions{
		GetURLOptions: GetURLOptions{
			GetOptions: GetOptions{
				CommonOptions: commonOpts,
			},
		},
	}
	cmd := &cobra.Command{
		Use:     "console",
		Short:   "Opens the Jenkins console",
		Long:    console_long,
		Example: console_example,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			CheckErr(err)
		},
	}
	options.addConsoleFlags(cmd)
	return cmd
}

func (o *ConsoleOptions) addConsoleFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&o.OnlyViewURL, "url", "u", false, "Only displays and the URL and does not open the browser")
	cmd.Flags().BoolVarP(&o.ClassicMode, "classic", "", false, "Use the classic Jenkins skin instead of Blue Ocean")

	o.addGetUrlFlags(cmd)
	o.JenkinsSelector.AddFlags(cmd)
}

func (o *ConsoleOptions) Run() error {
	kubeClient, ns, err := o.KubeClientAndDevNamespace()
	if err != nil {
		return err
	}
	prow, err := o.IsProw()
	if err != nil {
		return err
	}
	if prow {
		o.JenkinsSelector.UseCustomJenkins = true
	}
	jenkinsServiceName, err := o.PickCustomJenkinsName(&o.JenkinsSelector, kubeClient, ns)
	if err != nil {
		return err
	}
	if jenkinsServiceName == "" && !prow {
		jenkinsServiceName = kube.ServiceJenkins
	}
	return o.Open(jenkinsServiceName, "Jenkins Console")
}

func (o *ConsoleOptions) Open(name string, label string) error {
	var err error
	svcURL := ""
	ns := o.Namespace
	if ns == "" && o.Environment != "" {
		ns, err = o.FindEnvironmentNamespace(o.Environment)
		if err != nil {
			return err
		}
	}
	if ns != "" {
		svcURL, err = o.FindServiceInNamespace(name, ns)
	} else {
		svcURL, err = o.FindService(name)
	}
	if err != nil && name != "" {
		log.Infof("If the app %s is running in a different environment you could try: %s\n", util.ColorInfo(name), util.ColorInfo("jx get applications"))
	}
	if err != nil {
		return err
	}
	fullURL := svcURL
	if name == "jenkins" {
		fullURL = o.urlForMode(svcURL)
	}
	if o.OnlyViewHost {
		host := util.URLToHostName(svcURL)
		fmt.Fprintf(o.Out, "%s: %s\n", label, util.ColorInfo(host))
	} else {
		fmt.Fprintf(o.Out, "%s: %s\n", label, util.ColorInfo(fullURL))
	}
	if !o.OnlyViewURL && !o.OnlyViewHost {
		browser.OpenURL(fullURL)
	}
	return nil
}

func (o *ConsoleOptions) urlForMode(url string) string {
	if o.ClassicMode {
		return url
	} else {
		return util.UrlJoin(url, BlueOceanPath)
	}
}
