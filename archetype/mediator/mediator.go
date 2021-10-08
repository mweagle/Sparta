package mediator

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	/*
		"github.com/aws/aws-lambda-go/blob/main/events"
		"github.com/aws/aws-lambda-go/events"
	*/)

type TestUser struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name" jsonschema:"title=the name,description=The name of a friend,example=joe,example=lucy,default=alex"`
	Friends     []int                  `json:"friends,omitempty" jsonschema_description:"The list of IDs, omitted when empty"`
	Tags        map[string]interface{} `json:"tags,omitempty" jsonschema_extras:"a=b,foo=bar,foo=bar1"`
	BirthDate   time.Time              `json:"birth_date,omitempty" jsonschema:"oneof_required=date"`
	YearOfBirth string                 `json:"year_of_birth,omitempty" jsonschema:"oneof_required=year"`
	Metadata    interface{}            `json:"metadata,omitempty" jsonschema:"oneof_type=string;array"`
	FavColor    string                 `json:"fav_color,omitempty" jsonschema:"enum=red,enum=green,enum=blue"`
}

type TestOutput struct {
}

/*
func ()
func () error
func (TIn) error
func () (TOut, error)
func (context.Context) error
func (context.Context, TIn) error
func (context.Context) (TOut, error)
func (context.Context, TIn) (TOut, error)
*/

/*
So the first entry needs to be the event that
kicks everything off...

func NewMediatorApp(handlers...)


A handler is either a typed handler of custom events
or triggered off a rule. The rule needs to be bound, so it
can't be from a function signature. :thinking:


*/
type mediatorHandler func(context.Context, interface{}) (interface{}, error)

type TypeOneEvent struct {
	Wat string
}
type TypeTwoEvent struct {
	Mama string
}

func TestStepOne(ctx context.Context, rule BoundRuleEvent) (*TypeOneEvent, error) {
	return nil, nil

}

func TestStep2(ctx context.Context, event TypeOneEvent) (*TypeTwoEvent, error) {

	return nil, nil
}

type BoundRuleEvent struct {
	events.CloudWatchEvent
}

func NewListenerRule(events.CloudWatchEvent) mediatorHandler {
	return func(ctx context.Context, input interface{}) (interface{}, error) {

		return nil, nil
	}
}
