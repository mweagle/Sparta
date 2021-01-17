// +build !lambdabinary

package sparta

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	broadcast "github.com/dustin/go-broadcast"
	tcell "github.com/gdamore/tcell/v2"
	spartaAWS "github.com/mweagle/Sparta/aws"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

const (
	broadcasterFunctionSelect = "functionSelect"
	broadcasterFileSubmit     = "fileSubmit"
)

////////////////////////////////////////////////////////////////////////////////
//
// Public
//

// ExploreWithInputFilter allows the caller to provide additional filters
// for the source files that will be used as inputs
func ExploreWithInputFilter(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api APIGateway,
	site *S3Site,
	inputExtensions []string,
	s3BucketName string,
	buildTags string,
	linkerFlags string,
	logger *zerolog.Logger) error {

	// We need to setup the log output view so that we have a writer for the
	// log output
	logDataView := tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true)
	logDataView.SetChangedFunc(func() {
		logDataView.ScrollToEnd()
	})
	logDataView.SetBorder(true).SetTitle("Output")
	colorWriter := tview.ANSIWriter(logDataView)
	newLogger := logger.Output(colorWriter).Level(logger.GetLevel())
	logger = &newLogger

	// Great - everybody get's an aws session
	awsSession := spartaAWS.NewSession(logger)
	// Go get the stack and put the ARNs in the list of things. For that
	// we need to get the stack resources...
	cfSvc := cloudformation.New(awsSession)
	input := &cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(serviceName),
	}
	stackResourceOutputs, stackResourceOutputsErr := cfSvc.DescribeStackResources(input)
	if stackResourceOutputsErr != nil {
		return stackResourceOutputsErr
	}

	// Load the settings
	settingsMap := loadSettings()

	// Make the channel map
	channelMap := make(map[string]broadcast.Broadcaster)
	channelMap[broadcasterFunctionSelect] = broadcast.NewBroadcaster(1)
	channelMap[broadcasterFileSubmit] = broadcast.NewBroadcaster(1)
	application := tview.NewApplication()

	// Setup the rest of them...
	focusTargets := []tview.Primitive{}
	dropdown, selectorFocusable := newFunctionSelector(awsSession,
		stackResourceOutputs.StackResources,
		application,
		lambdaAWSInfos,
		settingsMap,
		channelMap[broadcasterFunctionSelect],
		logger)
	eventDropdown, eventFocusable := newEventInputSelector(awsSession,
		application,
		lambdaAWSInfos,
		settingsMap,
		inputExtensions,
		channelMap[broadcasterFunctionSelect],
		logger)
	outputView, outputViewFocusable := newCloudWatchLogTailView(awsSession,
		application,
		lambdaAWSInfos,
		settingsMap,
		channelMap[broadcasterFunctionSelect],
		logger)

	// There are four primary views
	if selectorFocusable != nil {
		focusTargets = append(focusTargets, selectorFocusable...)
	}
	if eventFocusable != nil {
		focusTargets = append(focusTargets, eventFocusable...)
	}
	if logDataView != nil {
		focusTargets = append(focusTargets, logDataView)
	}
	if outputViewFocusable != nil {
		focusTargets = append(focusTargets, outputViewFocusable...)
	}

	// Make it easy and use a Flex layout...
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dropdown, 3, 0, true).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(eventDropdown, 0, 1, false).
			AddItem(logDataView, 0, 2, false), 0, 1, false).
		AddItem(outputView, 0, 1, false)

	// Run  it...
	application.SetRoot(flex, true).SetFocus(flex)
	currentIndex := 0
	application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		logger.Debug().
			Interface("Key", event.Key()).
			Interface("Name", event.Name()).
			Interface("Modifier", event.Modifiers()).
			Msg("Input key")

		switch event.Key() {
		case tcell.KeyTab,
			tcell.KeyBacktab:

			direction := 1
			if event.Key() == tcell.KeyBacktab {
				direction = -1
			}
			nextIndex := (currentIndex + direction*1) % len(focusTargets)
			application.SetFocus(focusTargets[nextIndex])
			currentIndex = nextIndex
		default:
			// NOP
		}
		return event
	})
	return application.Run()

}

// Explore is an interactive command that brings up a GUI to test
// lambda functions previously deployed into AWS lambda. It's not supported in the
// AWS binary build
func Explore(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api APIGateway,
	site *S3Site,
	s3BucketName string,
	buildTags string,
	linkerFlags string,
	logger *zerolog.Logger) error {

	return ExploreWithInputFilter(serviceName,
		serviceDescription,
		lambdaAWSInfos,
		api,
		site,
		[]string{"json"},
		s3BucketName,
		buildTags,
		linkerFlags,
		logger)
}
