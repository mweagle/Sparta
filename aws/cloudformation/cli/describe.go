package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2Config "github.com/aws/aws-sdk-go-v2/config"
	awsv2CF "github.com/aws/aws-sdk-go-v2/service/cloudformation"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	validator "gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

/******************************************************************************/
// Global options
type optionsLinkStruct struct {
	StackName       string `validate:"required"`
	OutputDirectory string `validate:"required"`
}

var optionsLink optionsLinkStruct

// RootCmd represents the root Cobra command invoked for the discovery
// and serialization of an existing CloudFormation stack
var RootCmd = &cobra.Command{
	Use:   "link",
	Short: "Link is a tool to discover and serialize a prexisting CloudFormation stack",
	Long:  "",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		validateErr := validate.Struct(optionsLink)
		if nil != validateErr {
			return validateErr
		}
		// Make sure the output value is a directory
		osStat, osStatErr := os.Stat(optionsLink.OutputDirectory)
		if nil != osStatErr {
			return osStatErr
		}
		if !osStat.IsDir() {
			return errors.Errorf("--output (%s) is not a valid directory", optionsLink.OutputDirectory)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		runContext := context.Background()

		// Get the output and stuff it to a file
		awsConfig, awsConfigErr := awsv2Config.LoadDefaultConfig(runContext)
		if awsConfigErr != nil {
			return awsConfigErr
		}
		svc := awsv2CF.NewFromConfig(awsConfig)
		params := &awsv2CF.DescribeStacksInput{
			StackName: awsv2.String(optionsLink.StackName),
		}
		describeStacksResponse, describeStacksResponseErr := svc.DescribeStacks(runContext, params)

		if describeStacksResponseErr != nil {
			return describeStacksResponseErr
		}

		stackInfo, stackInfoErr := json.Marshal(describeStacksResponse)
		if stackInfoErr != nil {
			return errors.Wrapf(stackInfoErr, "Failed to describe stacks")
		}
		outputFilepath := filepath.Join(optionsLink.OutputDirectory, fmt.Sprintf("%s.json", optionsLink.StackName))
		err := ioutil.WriteFile(outputFilepath, stackInfo, 0600)
		if nil != err {
			return errors.Wrap(err, "Attempting to write output file")
		}
		fmt.Println("Created file: " + outputFilepath)
		fmt.Println(describeStacksResponse)
		return nil
	},
}

func init() {
	validate = validator.New()
	cobra.OnInitialize()
	RootCmd.PersistentFlags().StringVar(&optionsLink.StackName, "stackName", "", "CloudFormation Stack Name/ID to query")
	RootCmd.PersistentFlags().StringVar(&optionsLink.OutputDirectory, "output", "", "Output directory")
}

func main() {
	// Take a stack name and an output file...
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
