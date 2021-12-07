//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2CFTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	awsv2Lambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	broadcast "github.com/dustin/go-broadcast"
	tcell "github.com/gdamore/tcell/v2"
	prettyjson "github.com/hokaccha/go-prettyjson"
	spartaCWLogs "github.com/mweagle/Sparta/v3/aws/cloudwatch/logs"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

var (
	progressEmoji        = []string{"🌍", "🌎", "🌏"}
	windowsProgressEmoji = []string{"◐", "◓", "◑", "◒"}
)

////////////////////////////////////////////////////////////////////////////////
//
// Settings

var mu sync.Mutex

const (
	settingSelectedARN   = "functionARN"
	settingSelectedEvent = "selectedEvent"
)

func settingsFile() string {
	return filepath.Join(ScratchDirectory, "explore-settings.json")
}
func saveSetting(key string, value string) {
	settingsMap := loadSettings()
	settingsMap[string(key)] = value
	output, outputErr := json.MarshalIndent(settingsMap, "", " ")
	if outputErr != nil {
		return
	}
	mu.Lock()
	/* #nosec */
	writtenErr := ioutil.WriteFile(settingsFile(), output, os.ModePerm)
	if writtenErr != nil {
		fmt.Printf("Failed to save settings: %s", writtenErr.Error())
	}
	mu.Unlock()
}

func loadSettings() map[string]string {
	defaultSettings := make(map[string]string)
	settingsFile := settingsFile()
	mu.Lock()
	/* #nosec */
	bytes, bytesErr := ioutil.ReadFile(settingsFile)
	mu.Unlock()
	if bytesErr != nil {
		return defaultSettings
	}
	/* #nosec */
	umarshalErr := json.Unmarshal(bytes, &defaultSettings)
	if umarshalErr != nil {
		fmt.Printf("Failed to unmarshal: %s", umarshalErr.Error())
	}
	return defaultSettings
}

// Settings
//
////////////////////////////////////////////////////////////////////////////////

func writePrettyString(writer io.Writer, input string) {
	colorWriter := tview.ANSIWriter(writer)
	var jsonData map[string]interface{}
	jsonErr := json.Unmarshal([]byte(input), &jsonData)
	if jsonErr == nil {
		// pretty print it to colors...
		prettyString, prettyStringErr := prettyjson.Marshal(jsonData)
		if prettyStringErr == nil {
			/* #nosec */
			_, writeErr := io.WriteString(colorWriter, string(prettyString))
			if writeErr != nil {
				fmt.Printf("Failed to writeString: %s", writeErr.Error())
			}
		} else {
			/* #nosec */
			_, writeErr := io.WriteString(colorWriter, input)
			if writeErr != nil {
				fmt.Printf("Failed to writeString: %s", writeErr.Error())
			}
		}
	} else {
		/* #nosec */
		_, writeErr := io.WriteString(colorWriter, strings.TrimSpace(input))
		if writeErr != nil {
			fmt.Printf("Failed to writeString: %s", writeErr.Error())
		}
	}
	/* #nosec */
	_, writeErr := io.WriteString(writer, "\n")
	if writeErr != nil {
		fmt.Printf("Failed to writeString: %s", writeErr.Error())
	}
}

////////////////////////////////////////////////////////////////////////////////
//
// Select the function to test
//
func newFunctionSelector(awsConfig awsv2.Config,
	stackResources []awsv2CFTypes.StackResource,
	app *tview.Application,
	lambdaAWSInfos []*LambdaAWSInfo,
	settings map[string]string,
	onChangeBroadcaster broadcast.Broadcaster,
	logger *zerolog.Logger) (tview.Primitive, []tview.Primitive) {

	lambdaARN := func(stackID string, logicalName string) string {
		// stackID: arn:aws:cloudformation:us-west-2:123412341234:stack/MyHelloWorldStack-mweagle/54339e80-6686-11e8-90cd-503f20f2ad82
		// lambdaARN: arn:aws:lambda:us-west-2:123412341234:function:MyHelloWorldStack-mweagle_Hello_World
		stackParts := strings.Split(stackID, ":")
		lambdaARNParts := []string{
			"arn:aws:lambda:",
			stackParts[3],
			":",
			stackParts[4],
			":function:",
			logicalName,
		}
		return strings.Join(lambdaARNParts, "")
	}
	// Ok, walk the resources and assemble all the ARNs for the lambda functions
	lambdaFunctionARNs := []string{}
	for _, eachResource := range stackResources {
		if *eachResource.ResourceType == "AWS::Lambda::Function" {
			logger.Debug().
				Str("Resource", *eachResource.LogicalResourceId).
				Msg("Found provisioned Lambda function")
			lambdaFunctionARNs = append(lambdaFunctionARNs,
				lambdaARN(*eachResource.StackId, *eachResource.PhysicalResourceId))
		}
	}
	sort.Strings(lambdaFunctionARNs)
	selectedARN := settings[settingSelectedARN]
	selectedIndex := 0
	for index, eachARN := range lambdaFunctionARNs {
		if eachARN == selectedARN {
			selectedIndex = index
			break
		}
	}
	dropdown := tview.NewDropDown().
		SetCurrentOption(selectedIndex).
		SetLabel("Function ARN: ").
		SetOptions(lambdaFunctionARNs, nil)
	dropdown.SetBorder(true).SetTitle("Select Function")

	dropdownSelectFunc := func(text string, selectedIndex int) {
		if selectedIndex != -1 {
			saveSetting(settingSelectedARN, text)
			logger.Debug().Msgf("Selected: %s", selectedARN)
			onChangeBroadcaster.Submit(text)
		}
	}
	dropdown.SetSelectedFunc(dropdownSelectFunc)
	// Populate it...
	return dropdown, []tview.Primitive{dropdown}
}

////////////////////////////////////////////////////////////////////////////////
//
// Select the event to use to invoke the function
//
func newEventInputSelector(ctx context.Context,
	awsConfig awsv2.Config,
	app *tview.Application,
	lambdaAWSInfos []*LambdaAWSInfo,
	settings map[string]string,
	inputExtensionsFilters []string,
	functionSelectedBroadcaster broadcast.Broadcaster,
	logger *zerolog.Logger) (tview.Primitive, []tview.Primitive) {

	divider := strings.Repeat("━", 20)
	activeFunction := ""
	ch := make(chan interface{})
	functionSelectedBroadcaster.Register(ch)
	go func() {
		for {
			funcSelected := <-ch
			activeFunction = funcSelected.(string)
		}
	}()
	lambdaSvc := awsv2Lambda.NewFromConfig(awsConfig)

	// First walk the directory for anything that looks
	// like a JSON file...
	curDir, curDirErr := os.Getwd()
	if curDirErr != nil {
		return nil, nil
	}
	jsonFiles := []string{}
	walkerFunc := func(path string, info os.FileInfo, err error) error {
		for _, eachMatch := range inputExtensionsFilters {
			if strings.HasSuffix(strings.ToLower(filepath.Ext(path)), eachMatch) &&
				!strings.Contains(path, ScratchDirectory) {
				relPath := strings.TrimPrefix(path, curDir)
				jsonFiles = append(jsonFiles, relPath)
				logger.Debug().
					Str("RelativePath", relPath).
					Msg("Event file found")
			}
		}
		return nil
	}
	logger.Debug().
		Str("RelativePath", curDir).
		Msg("Walking directory")

	walkErr := filepath.Walk(curDir, walkerFunc)
	if walkErr != nil {
		logger.Error().
			Err(walkErr).
			Str("Directory", curDir).
			Msg("Failed to find JSON files in directory")
		return nil, nil
	}
	// Create all the views...
	var selectedJSONData []byte
	selectedInput := 0
	eventSelected := settings[settingSelectedEvent]
	for index, eachJSONFile := range jsonFiles {
		if eventSelected == eachJSONFile {
			selectedInput = index
			break
		}
	}
	eventDataView := tview.NewTextView().SetScrollable(true).SetDynamicColors(true)
	dropdown := tview.NewDropDown().
		SetCurrentOption(selectedInput).
		SetLabel("Event: ").
		SetOptions(jsonFiles, nil)

	selectEventData := func(text string, selectedIndex int) {
		// What's the selected item?
		if selectedIndex == -1 {
			return
		}
		logger.Debug().Str("EventInputPath", text).Msg("Event data source")
		eventDataView.Clear()

		// Save it...
		saveSetting(settingSelectedEvent, text)
		fullPath := curDir + text
		/* #nosec */
		jsonFile, jsonFileErr := ioutil.ReadFile(fullPath)
		if jsonFileErr != nil {
			writePrettyString(eventDataView, jsonFileErr.Error())
		} else {
			writePrettyString(eventDataView, string(jsonFile))
		}
		selectedJSONData = jsonFile
	}
	dropdown.SetSelectedFunc(selectEventData)
	submitButton := tview.NewButton("Submit")
	submitButton.SetBackgroundColorActivated(tcell.ColorDarkGreen)
	submitButton.SetLabelColorActivated(tcell.ColorWhite)
	submitButton.SetBackgroundColor(tcell.ColorGray)
	submitButton.SetLabelColor(tcell.ColorDarkGreen)
	submitButton.SetSelectedFunc(func() {
		logger.Debug().Str("ActiveFunction", activeFunction).Msg("Invoking function")
		// Submit it to lambda
		if activeFunction != "" {
			lambdaInput := &awsv2Lambda.InvokeInput{
				FunctionName: awsv2.String(activeFunction),
				Payload:      selectedJSONData,
			}
			invokeOutput, invokeOutputErr := lambdaSvc.Invoke(ctx, lambdaInput)
			if invokeOutputErr != nil {
				logger.Error().
					Err(invokeOutputErr).
					Msg("Failed to invoke Lambda function")
			} else if invokeOutput.FunctionError != nil {
				logger.Error().
					Str("Error", *invokeOutput.FunctionError).
					Msg("Lambda function produced an error")
			} else {
				var m interface{}

				jsonErr := json.Unmarshal(invokeOutput.Payload, &m)
				var responseData interface{}
				if jsonErr == nil {
					responseData = m
				} else {
					responseData = string(invokeOutput.Payload)
				}
				logger.Info().
					Interface("payload", responseData).
					Msg(divider + " AWS Lambda Response " + divider)
			}
		}
	})

	// Ok, so what we need now is a flexbox with a row,
	flexRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(dropdown, 0, 4, false).
		AddItem(submitButton, 10, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(flexRow, 1, 0, false).
		AddItem(eventDataView, 0, 1, false)
	flex.SetBorder(true).SetTitle("Select Event Input")
	return flex, []tview.Primitive{dropdown, submitButton, eventDataView}
}

////////////////////////////////////////////////////////////////////////////////
//
// Tail the cloudwatch logs for the active function
//
func newCloudWatchLogTailView(ctx context.Context,
	awsConfig awsv2.Config,
	app *tview.Application,
	lambdaAWSInfos []*LambdaAWSInfo,
	settings map[string]string,
	functionSelectedBroadcaster broadcast.Broadcaster,
	logger *zerolog.Logger) (tview.Primitive, []tview.Primitive) {

	osEmojiSet := progressEmoji
	switch runtime.GOOS {
	case "windows":
		osEmojiSet = windowsProgressEmoji
	}

	// Great - so what we need to do is listen for both the selected function
	// and a change in input. If we have values for both, then
	// go ahead and issue the request. We can do this with two
	// go-routines. The first one is just a go-routine that listens for cloudwatch log events
	// for the selected function.
	ch := make(chan interface{})
	functionSelectedBroadcaster.Register(ch)

	// So what we need here is a "Last event timestamp" entry and then the actual
	// content...
	cloudwatchLogInfoView := tview.NewTextView().SetDynamicColors(true)
	cloudwatchLogInfoView.SetBorder(true)
	logEventDataView := tview.NewTextView().SetDynamicColors(true)
	logEventDataView.SetScrollable(true)
	progressEmojiView := tview.NewTextView()

	// Ok, for this we need two colums, with the first column
	// being the
	flexView := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(cloudwatchLogInfoView, 0, 1, false), 3, 0, false).
		AddItem(logEventDataView, 0, 1, false).
		AddItem(progressEmojiView, 1, 0, false)
	flexView.SetBorder(true).SetTitle("CloudWatch Logs")

	updateCloudWatchLogInfoView := func(logGroupName string, latestTS int64) {
		// Ref: https://godoc.org/github.com/rivo/tview#hdr-Colors
		// Color tag definition: [<foreground>:<background>:<flags>]
		cloudwatchLogInfoView.Clear()
		ts := ""
		if latestTS != 0 {
			ts = time.Unix(latestTS, 0).Format(time.RFC3339)
		}
		msg := fmt.Sprintf("[-:-:b]LogGroupName[-:-:-]: [-:-:d]%s",
			logGroupName)
		if ts != "" {
			msg += fmt.Sprintf(" ([-:-:b]Latest Event[-:-:-]: [-:-:d]%s)", ts)
		}
		writePrettyString(cloudwatchLogInfoView, msg)
	}
	updateCloudWatchLogInfoView("", 0)
	// When we get a new function then
	var selectedFunction string
	go func() {
		var doneChan chan bool
		var ticker *time.Ticker
		lastTime := int64(0)
		animationIndex := 0

		for {
			funcSelected := <-ch
			if selectedFunction == funcSelected.(string) {
				continue
			}
			selectedFunction = funcSelected.(string)
			logEventDataView.Clear()
			if doneChan != nil {
				doneChan <- true
				progressEmojiView.Clear()
			}
			if ticker != nil {
				ticker.Stop()
			}
			ticker = time.NewTicker(time.Millisecond * 333)
			lambdaARN := selectedFunction
			lambdaParts := strings.Split(lambdaARN, ":")
			logGroupName := fmt.Sprintf("/aws/lambda/%s", lambdaParts[len(lambdaParts)-1])
			logger.Debug().
				Str("Name", logGroupName).
				Msg("CloudWatch LogGroupName")

			// Put this as the label in the view...
			doneChan = make(chan bool)
			messages := spartaCWLogs.TailWithContext(ctx,
				doneChan,
				awsConfig,
				logGroupName,
				"",
				logger)
			// Go read it...
			go func() {
				for {
					select {
					case event := <-messages:
						{
							lastTime = *event.Timestamp / 1000
							updateCloudWatchLogInfoView(logGroupName, lastTime)
							writePrettyString(logEventDataView, *event.Message)
							logger.Debug().
								Str("EventID", *event.EventId).
								Msg("Event received")
							logEventDataView.ScrollToEnd()
							app.Draw()
						}
					case <-ticker.C:
						/* #nosec */
						animationIndex = (animationIndex + 1) % len(osEmojiSet)
						progressEmojiView.Clear()
						progressText := fmt.Sprintf("%s Waiting for events...", osEmojiSet[animationIndex])
						/* #nosec */
						_, writeErr := io.WriteString(progressEmojiView, progressText)
						if writeErr != nil {
							fmt.Printf("Failed to write string: %s", writeErr.Error())
						}
						// Update the other stuff
						updateCloudWatchLogInfoView(logGroupName, lastTime)
						app.Draw()
					}
				}
			}()
		}
	}()
	return flexView, []tview.Primitive{logEventDataView}
}
