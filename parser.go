package vera

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// type intermediateSignalParsing struct {
// 	name         string
// 	bitInfo      string
// 	factorOffset string
// 	minMax       string
// 	unit         string
// 	receivers    string
// }

func Parse(r io.Reader) (*Config, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	config := &Config{}

	content := string(bytes)
	content = replaceNewLineCharacters(content)

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return config, nil
	}

	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "BO_") {
			messageInstruction := lines[i]

			for j := i + 1; j < len(lines); j++ {
				line := strings.TrimFunc(lines[j], func(r rune) bool {
					return r == ' ' || r == '\t'
				})
				if !strings.HasPrefix(line, "SG_") {
					i = j - 1
					break
				}

				messageInstruction += "\n" + line
			}

			messageConfig, err := parseMessageInstruction(messageInstruction)
			if err != nil {
				return nil, err
			}

			config.Messages = append(config.Messages, *messageConfig)
		} else if strings.HasPrefix(lines[i], "TP_") {
			signalTopic, err := parseSignalTopic(lines[i])
			if err != nil {
				return nil, err
			}

			config.Topics = append(config.Topics, *signalTopic)
		}
	}

	return config, nil
}

func parseMessageInstruction(messageInstruction string) (*Message, error) {
	if !strings.HasPrefix(messageInstruction, "BO_") {
		return nil, ErrorMessageWrongPrefix
	}

	messageLines := strings.Split(messageInstruction, "\n")
	if len(messageLines) == 0 {
		return nil, ErrorMessageZeroLines
	}

	messageDefinition := messageLines[0]
	message, err := parseMessageDefinition(messageDefinition)
	if err != nil {
		return nil, err
	}

	if len(messageLines) == 1 {
		return message, nil
	}

	signals, err := parseMessageSignals(strings.Join(messageLines[1:], "\n"))
	if err != nil {
		return nil, err
	}
	message.Signals = signals

	return message, nil
}

func parseMessageDefinition(messageDefinition string) (*Message, error) {
	messageDefinitionParts := strings.Fields(messageDefinition)
	if len(messageDefinitionParts) != 5 {
		return nil, ErrorMessageDefinitionWrongStructure
	}

	messageIDStr := messageDefinitionParts[1]
	messageID, err := parseMessageID(messageIDStr)
	if err != nil {
		return nil, err
	}

	messageName := messageDefinitionParts[2]
	if !strings.HasSuffix(messageName, ":") {
		return nil, ErrorMessageNameEndsNotWithColon
	}
	messageName = strings.TrimSuffix(messageName, ":")

	dlcStr := messageDefinitionParts[3]
	dlc, err := strconv.Atoi(dlcStr)
	if err != nil {
		return nil, ErrorMessageDLCNotInteger
	}

	transmitter := messageDefinitionParts[4]

	message := &Message{
		ID:          messageID,
		Name:        messageName,
		DLC:         uint8(dlc),
		Transmitter: Node(transmitter),
	}

	return message, nil
}

func parseMessageID(messageIDStr string) (uint32, error) {
	if !strings.HasPrefix(messageIDStr, "0x") {
		messageID, err := strconv.Atoi(messageIDStr)
		if err != nil {
			return 0, ErrorMessageIDNotInteger
		}

		return uint32(messageID), nil
	}

	messageIDHex := strings.TrimPrefix(messageIDStr, "0x")
	messageID, err := strconv.ParseUint(messageIDHex, 16, 32)
	if err != nil {
		return 0, ErrorMessageIDNotInteger
	}

	return uint32(messageID), nil
}

func parseMessageSignals(messageSignals string) ([]Signal, error) {
	signalsLines := strings.Split(messageSignals, "\n")
	signalsSlice := make([]Signal, len(signalsLines))

	if len(signalsSlice) == 0 {
		return signalsSlice, nil
	}

	for i, line := range signalsLines {
		if !strings.HasPrefix(line, "SG_") {
			return nil, fmt.Errorf("signal line %d doesn't start with 'SG_'", i)
		}

		lineParts := strings.Fields(line)
		if len(lineParts) < 7 {
			return nil, fmt.Errorf(`signal line %d is not well structured, must adhere to:
SG_ <SignalName> : <StartBit>|<Length>@<BitOrder><Signed> (<Factor>,<Offset>) [<Min>,<Max>] "<Unit>" <...Receivers>`, i)
		}

		if lineParts[2] != ":" {
			return nil, fmt.Errorf("signal line %d has not a ':' between <SignalName> and <StartBit>", i)
		}

		signal := Signal{
			Name: lineParts[1],
		}

		signalBitInfo := lineParts[3]
		if err := parseMessageBitInfo(&signal, i, signalBitInfo); err != nil {
			return nil, err
		}

		signalFactorOffset := lineParts[4]
		if err := parseMessageFactorOffset(&signal, i, signalFactorOffset); err != nil {
			return nil, err
		}

		signalMinMax := lineParts[5]
		if err := parseMessageMinMax(&signal, i, signalMinMax); err != nil {
			return nil, err
		}

		signalUnitStr := lineParts[6]
		if err := parseMessageUnit(&signal, i, signalUnitStr); err != nil {
			return nil, err
		}

		parseMessageReceivers(&signal, lineParts[7])

		signalsSlice[i] = signal
	}

	return signalsSlice, nil
}

func parseMessageBitInfo(signal *Signal, i int, signalBitInfo string) error {
	signalBitFirstSplit := strings.Split(signalBitInfo, "@")
	if len(signalBitFirstSplit) != 2 {
		return fmt.Errorf("signal line %d has invalid bit info: %s", i, signalBitInfo)
	}

	signalStartBitAndLength := strings.Split(signalBitFirstSplit[0], "|")
	if len(signalStartBitAndLength) != 2 {
		return fmt.Errorf("signal line %d has invalid start bit and length: %s", i, signalBitFirstSplit[0])
	}
	signalStartBit, err := strconv.Atoi(signalStartBitAndLength[0])
	if err != nil {
		return fmt.Errorf("signal line %d has invalid start bit: %s", i, signalStartBitAndLength[0])
	}
	signalLength, err := strconv.Atoi(signalStartBitAndLength[1])
	if err != nil {
		return fmt.Errorf("signal line %d has invalid length: %s", i, signalStartBitAndLength[1])
	}

	signalOtherInfo := signalBitFirstSplit[1]
	if len(signalOtherInfo) < 2 {
		return fmt.Errorf("signal line %d has invalid bit order and signed: %s", i, signalOtherInfo)
	}

	signalEndianness, err := strconv.ParseUint(string(signalOtherInfo[0]), 10, 1)
	if err != nil {
		return fmt.Errorf("signal line %d has invalid bit order: %s", i, string(signalOtherInfo[0]))
	}
	var signalSigned bool
	switch signalOtherInfo[1] {
	case '+':
		signalSigned = false
	case '-':
		signalSigned = true
	default:
		return fmt.Errorf("signal line %d has invalid signed: %s", i, string(signalOtherInfo[1]))
	}

	signal.Endianness = Endianness(signalEndianness)
	signal.Length = uint8(signalLength)
	signal.Signed = signalSigned
	signal.StartBit = uint8(signalStartBit)

	if len(signalOtherInfo) == 2 {
		return nil
	}

	signalDecimalFormat := signalOtherInfo[2:]
	if signalDecimalFormat[0] != '(' || signalDecimalFormat[len(signalDecimalFormat)-1] != ')' {
		return fmt.Errorf("signal line %d has invalid decimal format: %s", i, signalDecimalFormat)
	}

	signalDecimalFormat = strings.TrimPrefix(strings.TrimSuffix(signalDecimalFormat, ")"), "(")
	signalDecimalFormatParts := strings.Split(signalDecimalFormat, ",")
	if len(signalDecimalFormatParts) != 2 {
		return fmt.Errorf("signal line %d has invalid decimal format: %s", i, signalDecimalFormat)
	}

	integerFiguresStr := signalDecimalFormatParts[0]
	decimalFiguresStr := signalDecimalFormatParts[1]

	integerFigures, err := strconv.Atoi(integerFiguresStr)
	if err != nil {
		return fmt.Errorf("signal line %d has invalid decimal format: %s", i, signalDecimalFormat)
	}
	decimalFigures, err := strconv.Atoi(decimalFiguresStr)
	if err != nil {
		return fmt.Errorf("signal line %d has invalid decimal format: %s", i, signalDecimalFormat)
	}

	signal.IntegerFigures = uint8(integerFigures)
	signal.DecimalFigures = uint8(decimalFigures)

	return nil
}

func parseMessageFactorOffset(signal *Signal, i int, signalFactorOffset string) error {
	if !strings.HasPrefix(signalFactorOffset, "(") || !strings.HasSuffix(signalFactorOffset, ")") {
		return fmt.Errorf("signal line %d has invalid factor and offset: %s", i, signalFactorOffset)
	}
	signalFactorOffsetParts := strings.Split(signalFactorOffset[1:len(signalFactorOffset)-1], ",")
	if len(signalFactorOffsetParts) != 2 {
		return fmt.Errorf("signal line %d has invalid factor and offset: %s", i, signalFactorOffset)
	}

	signalFactor, err := strconv.ParseFloat(signalFactorOffsetParts[0], 32)
	if err != nil {
		return fmt.Errorf("signal line %d has invalid factor: %s", i, string(signalFactorOffsetParts[0]))
	}
	signalOffset, err := strconv.ParseFloat(signalFactorOffsetParts[1], 32)
	if err != nil {
		return fmt.Errorf("signal line %d has invalid offset: %s", i, string(signalFactorOffsetParts[1]))
	}

	signal.Factor = float32(signalFactor)
	signal.Offset = float32(signalOffset)

	return nil
}

func parseMessageMinMax(signal *Signal, i int, signalMinMax string) error {
	if !strings.HasPrefix(signalMinMax, "[") || !strings.HasSuffix(signalMinMax, "]") {
		return fmt.Errorf("signal line %d has invalid min/max: %s", i, signalMinMax)
	}
	signalMinMaxParts := strings.Split(signalMinMax[1:len(signalMinMax)-1], "|")
	if len(signalMinMaxParts) != 2 {
		return fmt.Errorf("signal line %d has invalid min/max: %s", i, signalMinMax)
	}
	signalMin, err := strconv.ParseFloat(signalMinMaxParts[0], 32)
	if err != nil {
		return fmt.Errorf("signal line %d has invalid min: %s", i, string(signalMinMaxParts[0]))
	}
	signalMax, err := strconv.ParseFloat(signalMinMaxParts[1], 32)
	if err != nil {
		return fmt.Errorf("signal line %d has invalid max: %s", i, string(signalMinMaxParts[1]))
	}

	signal.Min = float32(signalMin)
	signal.Max = float32(signalMax)

	return nil
}

func parseMessageUnit(signal *Signal, i int, signalUnitStr string) error {
	if !strings.HasPrefix(signalUnitStr, "\"") || !strings.HasSuffix(signalUnitStr, "\"") {
		return fmt.Errorf("signal line %d has invalid unit: %s", i, signalUnitStr)
	}
	signal.Unit = signalUnitStr[1 : len(signalUnitStr)-1]

	return nil
}

func parseMessageReceivers(signal *Signal, signalReceivers string) {
	receiversStr := strings.Split(signalReceivers, ",")

	for _, r := range receiversStr {
		signal.Receivers = append(signal.Receivers, Node(r))
	}
}

func parseSignalTopic(topicLine string) (*SignalTopic, error) {
	lineParts := strings.Fields(topicLine)
	if len(lineParts) != 3 {
		return nil, fmt.Errorf(`signal topic has wrong structure: %s
Should be:
	TP_ <SignalName> <Topic>`, topicLine)
	}

	signalName := lineParts[1]
	topic := lineParts[2]

	signalTopic := &SignalTopic{
		Topic:  topic,
		Signal: signalName,
	}

	return signalTopic, nil
}
