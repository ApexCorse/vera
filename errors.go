package vera

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

var (
	ErrorMessageDLCOutOfBounds              = errors.New("message DLC must be a number from 1 to 8")
	ErrorSignalStartBitOutOfBounds          = errors.New("signal start bit must be a number from 0 to 63")
	ErrorSignalLengthOutOfBounds            = errors.New("signal length must be a number from 1 to 64")
	ErrorSignalLengthsGreaterThanMessageDLC = errors.New("sum of signal lengths is greater than message DLC")
	ErrorSignalFactorIsZero                 = errors.New("signal factor cannot be zero")
	ErrorInvalidSignalTopic                 = errors.New("signal topic is invalid")
)
