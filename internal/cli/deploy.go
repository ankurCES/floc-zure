package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ankurCES/floc-zure/internal/arm"
	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "ARM template deployment commands",
	Long:  "Deploy Azure resources from ARM template JSON files using the simulator or real Azure CLI.",
}

var deployCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Deploy an ARM template",
	RunE: func(cmd *cobra.Command, args []string) error {
		templatePath, _ := cmd.Flags().GetString("template")
		paramsPath, _ := cmd.Flags().GetString("parameters")
		name, _ := cmd.Flags().GetString("name")
		rg, _ := cmd.Flags().GetString("resource-group")

		if templatePath == "" {
			return fmt.Errorf("--template is required")
		}
		if name == "" {
			name = fmt.Sprintf("deploy-%d", time.Now().Unix())
		}
		if rg == "" {
			rg = "arm-deployed"
		}

		// Parse template
		tmpl, err := arm.ParseTemplate(templatePath)
		if err != nil {
			return fmt.Errorf("parse template: %w", err)
		}

		// Parse parameters
		var suppliedParams map[string]interface{}
		if paramsPath != "" {
			suppliedParams, err = arm.ParseParameterFile(paramsPath)
			if err != nil {
				return fmt.Errorf("parse parameters: %w", err)
			}
		}

		// Resolve
		result, err := arm.Resolve(tmpl, suppliedParams)
		if err != nil {
			return fmt.Errorf("resolve template: %w", err)
		}

		// Deploy
		exec := azure.NewCLIExecutorImpl()
		deployer := arm.NewDeployer(exec)
		dep, err := deployer.Deploy(cmd.Context(), result, name, rg)
		if err != nil {
			// Still print the partial deployment
			data, _ := json.MarshalIndent(dep, "", "  ")
			fmt.Fprintln(os.Stderr, string(data))
			return err
		}

		data, _ := json.MarshalIndent(dep, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var deployValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate an ARM template without deploying",
	RunE: func(cmd *cobra.Command, args []string) error {
		templatePath, _ := cmd.Flags().GetString("template")
		paramsPath, _ := cmd.Flags().GetString("parameters")
		rg, _ := cmd.Flags().GetString("resource-group")

		if templatePath == "" {
			return fmt.Errorf("--template is required")
		}
		if rg == "" {
			rg = "arm-deployed"
		}

		tmpl, err := arm.ParseTemplate(templatePath)
		if err != nil {
			return fmt.Errorf("parse template: %w", err)
		}

		var suppliedParams map[string]interface{}
		if paramsPath != "" {
			suppliedParams, err = arm.ParseParameterFile(paramsPath)
			if err != nil {
				return fmt.Errorf("parse parameters: %w", err)
			}
		}

		result, err := arm.Resolve(tmpl, suppliedParams)
		if err != nil {
			return fmt.Errorf("resolve template: %w", err)
		}

		exec := azure.NewCLIExecutorImpl()
		deployer := arm.NewDeployer(exec)
		if err := deployer.Validate(result, rg); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		fmt.Println("Template is valid.")
		return nil
	},
}

var deployShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a deployment result file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			return fmt.Errorf("--file is required")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read deployment file: %w", err)
		}
		var dep arm.Deployment
		if err := json.Unmarshal(data, &dep); err != nil {
			return fmt.Errorf("parse deployment JSON: %w", err)
		}
		out, _ := json.MarshalIndent(dep, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	deployCreateCmd.Flags().String("template", "", "Path to ARM template JSON")
	deployCreateCmd.Flags().String("parameters", "", "Path to parameters JSON")
	deployCreateCmd.Flags().String("name", "", "Deployment name")
	deployCreateCmd.Flags().String("resource-group", "", "Target resource group")

	deployValidateCmd.Flags().String("template", "", "Path to ARM template JSON")
	deployValidateCmd.Flags().String("parameters", "", "Path to parameters JSON")
	deployValidateCmd.Flags().String("resource-group", "", "Target resource group")

	deployShowCmd.Flags().String("file", "", "Path to deployment result JSON")

	deployCmd.AddCommand(deployCreateCmd, deployValidateCmd, deployShowCmd)
	rootCmd.AddCommand(deployCmd)
}
