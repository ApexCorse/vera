package parser

import "errors"

var (
	ErrorMessageWrongPrefix              = errors.New("message does not start with 'BO_'")
	ErrorMessageZeroLines                = errors.New("message has zero lines somehow")
	ErrorMessageDefinitionWrongStructure = errors.New(`message definition must be composed of 5 elements:
BO_ <MessageID> <MessageName>: <DLC> <TransmitterNode>`)
	ErrorMessageIDNotInteger         = errors.New("message ID must be an base 10 or hexadecimal integer")
	ErrorMessageNameEndsNotWithColon = errors.New("message name must end with a ':'")
	ErrorMessageDLCNotInteger        = errors.New("message DLC must be a base 10 integer")
)
