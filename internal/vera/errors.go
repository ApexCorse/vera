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
	ErrorMessageLengthOutOfBounds              = errors.New("message DLC must be a number from 1 to 8")
	ErrorSignalLengthOutOfBounds               = errors.New("signal start byte must be a number from 1 to 8")
	ErrorSignalStartByteOutOfBounds            = errors.New("signal DLC must be a number from 1 to 8")
	ErrorSignalLengthsGreaterThanMessageLegnth = errors.New("sum of signal lengths is greater than message length")
	ErrorSignalFactorIsZero                    = errors.New("signal factor cannot be zero")
	ErrorSignalFigures                         = errors.New("sum of integer and decimal figures must be equal to Length * 8")
)
