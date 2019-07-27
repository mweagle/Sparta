// +build !lambdabinary

package sparta

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	broadcast "github.com/dustin/go-broadcast"
	"github.com/gdamore/tcell"
	spartaAWS "github.com/mweagle/Sparta/aws"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

const (
	broadcasterFunctionSelect = "functionSelect"
	broadcasterFileSubmit     = "fileSubmit"
)

////////////////////////////////////////////////////////////////////////////////
//
// Public
//

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
	logger *logrus.Logger) error {

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

	// That's the list of functions, which we can map up against the operations to perform
	// Create the logger first, since it will change the output sink
	logView, logViewFocusable := newLogOutputView(awsSession,
		application,
		lambdaAWSInfos,
		settingsMap,
		logger)

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
	if logViewFocusable != nil {
		focusTargets = append(focusTargets, logViewFocusable...)
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
			AddItem(logView, 0, 2, false), 0, 1, false).
		AddItem(outputView, 0, 1, false)

	// Run  it...
	application.SetRoot(flex, true).SetFocus(flex)
	currentIndex := 0
	application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		logger.WithFields(logrus.Fields{
			"Key":      event.Key(),
			"Name":     event.Name(),
			"Modifier": event.Modifiers(),
		}).Debug("Input key")

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
