package sparta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/lambda"
	broadcast "github.com/dustin/go-broadcast"
	"github.com/gdamore/tcell"
	"github.com/hokaccha/go-prettyjson"
	spartaCWLogs "github.com/mweagle/Sparta/aws/cloudwatchlogs"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

var (
	progressEmoji = []string{"🌍", "🌎", "🌏"}
)

func writePrettyString(writer io.Writer, input string) {
	colorWriter := tview.ANSIIWriter(writer)
	var jsonData map[string]interface{}
	jsonErr := json.Unmarshal([]byte(input), &jsonData)
	if jsonErr == nil {
		// pretty print it to colors...
		prettyString, prettyStringErr := prettyjson.Marshal(jsonData)
		if prettyStringErr == nil {
			/* #nosec */
			io.WriteString(colorWriter, string(prettyString))
		} else {
			/* #nosec */
			io.WriteString(colorWriter, input)
		}
	} else {
		/* #nosec */

		io.WriteString(colorWriter, fmt.Sprintf("%s", strings.TrimSpace(input)))
	}
	/* #nosec */
	io.WriteString(writer, "\n")
}

////////////////////////////////////////////////////////////////////////////////
//
// Select the function to test
//
func newFunctionSelector(awsSession *session.Session,
	stackResources []*cloudformation.StackResource,
	app *tview.Application,
	lambdaAWSInfos []*LambdaAWSInfo,
	onChangeBroadcaster broadcast.Broadcaster,
	logger *logrus.Logger) (tview.Primitive, []tview.Primitive) {

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
			logger.WithField("Resource", *eachResource.LogicalResourceId).Debug("Found provisioned Lambda function")
			lambdaFunctionARNs = append(lambdaFunctionARNs, lambdaARN(*eachResource.StackId, *eachResource.PhysicalResourceId))
		}
	}
	sort.Strings(lambdaFunctionARNs)
	dropdown := tview.NewDropDown().
		SetCurrentOption(0).
		SetLabel("Function ARN: ").
		SetOptions(lambdaFunctionARNs, nil)
	dropdown.SetBorder(true).SetTitle("Select Function")

	dropdownDoneFunc := func(key tcell.Key) {
		selectedIndex, value := dropdown.GetCurrentOption()
		if selectedIndex != -1 {
			onChangeBroadcaster.Submit(value)
		}
	}
	dropdown.SetDoneFunc(dropdownDoneFunc)
	// Populate it...
	dropdownDoneFunc(tcell.KeyEnter)
	return dropdown, []tview.Primitive{dropdown}
}

////////////////////////////////////////////////////////////////////////////////
//
// Select the event to use to invoke the function
//
func newEventInputSelector(awsSession *session.Session,
	app *tview.Application,
	lambdaAWSInfos []*LambdaAWSInfo,
	functionSelectedBroadcaster broadcast.Broadcaster,
	logger *logrus.Logger) (tview.Primitive, []tview.Primitive) {

	activeFunction := ""
	ch := make(chan interface{})
	functionSelectedBroadcaster.Register(ch)
	go func() {
		for {
			select {
			case funcSelected := <-ch:
				activeFunction = funcSelected.(string)
			}
		}
	}()
	lambdaSvc := lambda.New(awsSession)

	// First walk the directory for anything that looks
	// like a JSON file...
	curDir, curDirErr := os.Getwd()
	if curDirErr != nil {
		return nil, nil
	}
	jsonFiles := []string{}
	walkerFunc := func(path string, info os.FileInfo, err error) error {
		if strings.ToLower(filepath.Ext(path)) == ".json" &&
			!strings.Contains(path, ScratchDirectory) {
			relPath := strings.TrimPrefix(path, curDir)
			jsonFiles = append(jsonFiles, relPath)
			logger.WithField("RelativePath", relPath).Debug("Event file found")
		}
		return nil
	}
	walkErr := filepath.Walk(curDir, walkerFunc)
	if walkErr != nil {
		logger.WithError(walkErr).Error("Failed to find JSON files in directory: " + curDir)
		return nil, nil
	}
	// Create all the views...
	var selectedJSONData []byte

	eventDataView := tview.NewTextView().SetScrollable(true).SetDynamicColors(true)
	dropdown := tview.NewDropDown().
		SetCurrentOption(0).
		SetLabel("Event: ").
		SetOptions(jsonFiles, nil)

	submitEventData := func(key tcell.Key) {
		// What's the selected item?
		selected, value := dropdown.GetCurrentOption()
		if selected == -1 {
			return
		}
		eventDataView.Clear()
		fullPath := curDir + value
		/* #nosec */
		jsonFile, jsonFileErr := ioutil.ReadFile(fullPath)
		if jsonFileErr != nil {
			writePrettyString(eventDataView, jsonFileErr.Error())
		} else {
			writePrettyString(eventDataView, string(jsonFile))
		}
		selectedJSONData = jsonFile
	}
	submitEventData(tcell.KeyEnter)
	dropdown.SetDoneFunc(submitEventData)
	submitButton := tview.NewButton("Submit")
	submitButton.SetBackgroundColorActivated(tcell.ColorDarkGreen)
	submitButton.SetLabelColorActivated(tcell.ColorWhite)
	submitButton.SetBackgroundColor(tcell.ColorGray)
	submitButton.SetLabelColor(tcell.ColorDarkGreen)
	submitButton.SetSelectedFunc(func() {
		if activeFunction == "" {
			return
		}
		// Submit it to lambda
		if activeFunction != "" {
			lambdaInput := &lambda.InvokeInput{
				FunctionName: aws.String(activeFunction),
				Payload:      selectedJSONData,
			}
			invokeOutput, invokeOutputErr := lambdaSvc.Invoke(lambdaInput)
			if invokeOutputErr != nil {
				logger.WithFields(logrus.Fields{
					"Error": invokeOutputErr,
				}).Error("Failed to invoke Lambda function")
			} else if invokeOutput.FunctionError != nil {
				logger.WithFields(logrus.Fields{
					"Error": invokeOutput.FunctionError,
				}).Error("Lambda function produced an error")
			} else {
				var m interface{}
				var entry *logrus.Entry
				jsonErr := json.Unmarshal(invokeOutput.Payload, &m)
				if jsonErr == nil {
					entry = logger.WithField("Body", m)
				} else {
					entry = logger.WithField("Body", string(invokeOutput.Payload))
				}
				entry.Info("AWS Lambda Response")
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
func newCloudWatchLogTailView(awsSession *session.Session,
	app *tview.Application,
	lambdaAWSInfos []*LambdaAWSInfo,
	functionSelectedBroadcaster broadcast.Broadcaster,
	logger *logrus.Logger) (tview.Primitive, []tview.Primitive) {

	// Great - so what we need to do is listen for both the selected function
	// and a change in input. If we have values for both, then
	// go ahead and issue the request. We can do this with two
	// go-routines. The first one is just a go-routine that listens for cloudwatch log events
	// for the selected function. TODO - filter
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
			AddItem(cloudwatchLogInfoView, 0, 1, false), 4, 0, false).
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

		msg := fmt.Sprintf("[-:-:b]LogGroupName[-:-:-]: [-:-:d]%s\n[-:-:b]Latest Event[-:-:-]: [-:-:d]%s",
			logGroupName,
			ts)
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
			select {
			case funcSelected := <-ch:
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
				logger.WithField("Name", logGroupName).Debug("CloudWatch LogGroupName")

				// Put this as the label in the view...
				doneChan = make(chan bool)
				messages := spartaCWLogs.TailWithContext(context.Background(),
					doneChan,
					awsSession,
					logGroupName,
					"",
					logger)
				// Go read it...
				go func() {
					for {
						select {
						case event := <-messages:
							{
								logger.WithField("EventID", *event.EventId).Debug("Event received")
								lastTime = *event.Timestamp / 1000
								updateCloudWatchLogInfoView(logGroupName, lastTime)
								writePrettyString(logEventDataView, *event.Message)
								logEventDataView.ScrollToEnd()
								app.Draw()
							}
						case <-ticker.C:
							/* #nosec */
							animationIndex = (animationIndex + 1) % len(progressEmoji)
							progressEmojiView.Clear()
							progressText := fmt.Sprintf("%s Waiting for events...", progressEmoji[animationIndex])
							/* #nosec */
							io.WriteString(progressEmojiView, progressText)
							// Update the other stuff
							updateCloudWatchLogInfoView(logGroupName, lastTime)
							app.Draw()
						}
					}
				}()
			}
		}
	}()
	return flexView, []tview.Primitive{logEventDataView}
}

////////////////////////////////////////////////////////////////////////////////
//
// Redirect the logger to the log view
//
func newLogOutputView(awsSession *session.Session,
	app *tview.Application,
	lambdaAWSInfos []*LambdaAWSInfo,
	logger *logrus.Logger) (tview.Primitive, []tview.Primitive) {

	logDataView := tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true)
	logDataView.SetChangedFunc(func() {
		logDataView.ScrollToEnd()
	})
	logDataView.SetBorder(true).SetTitle("Log")

	colorWriter := tview.ANSIIWriter(logDataView)
	logger.Out = colorWriter
	return logDataView, []tview.Primitive{logDataView}
}
