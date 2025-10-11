package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

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
		if !strings.HasPrefix(lines[i], "BO_") {
			continue
		}
		messageInstruction := lines[i]

		for j := i + 1; j < len(lines); j++ {
			if !strings.HasPrefix(lines[j], "\tSG_") {
				i = j - 1
				break
			}

			messageInstruction += "\n" + lines[j]
		}

		messageConfig, err := parseMessageInstruction(messageInstruction)
		if err != nil {
			return nil, err
		}

		config.Messages = append(config.Messages, *messageConfig)
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
		Length:      uint8(dlc),
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
		if !strings.HasPrefix(line, "\tSG_") {
			return nil, fmt.Errorf("signal line %d is not indented or doesn't start with 'SG_'", i)
		}
		line = strings.TrimPrefix(line, "\t")

		lineParts := strings.Fields(line)
		if len(lineParts) < 7 {
			return nil, fmt.Errorf(`signal line %d is not well structured, must adhere to:
SG_ <SignalName> : <StartByte>|<Length>@<ByteOrder><Signed> (<Factor>,<Offset>) [<Min>,<Max>] "<Unit>" <...Receivers>`, i)
		}

		if lineParts[2] != ":" {
			return nil, fmt.Errorf("signal line %d has not a ':' between <SignalName> and <StartByte>", i)
		}

		signal := Signal{
			Name: lineParts[1],
		}

		signalBytesInfo := lineParts[3]
		if err := parseMessageBytesInfo(&signal, i, signalBytesInfo); err != nil {
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

func parseMessageBytesInfo(signal *Signal, i int, signalBytesInfo string) error {
	signalBytesFirstSplit := strings.Split(signalBytesInfo, "@")
	if len(signalBytesFirstSplit) != 2 {
		return fmt.Errorf("signal line %d has invalid bytes info: %s", i, signalBytesInfo)
	}

	signalStartByteAndLength := strings.Split(signalBytesFirstSplit[0], "|")
	if len(signalStartByteAndLength) != 2 {
		return fmt.Errorf("signal line %d has invalid start byte and length: %s", i, signalBytesFirstSplit[0])
	}
	signalStartByte, err := strconv.Atoi(signalStartByteAndLength[0])
	if err != nil {
		return fmt.Errorf("signal line %d has invalid start byte: %s", i, signalStartByteAndLength[0])
	}
	signalLength, err := strconv.Atoi(signalStartByteAndLength[1])
	if err != nil {
		return fmt.Errorf("signal line %d has invalid length: %s", i, signalStartByteAndLength[1])
	}

	signalEndiannessAndSigned := signalBytesFirstSplit[1]
	if len(signalEndiannessAndSigned) != 2 {
		return fmt.Errorf("signal line %d has invalid byte order and signed: %s", i, signalEndiannessAndSigned)
	}

	signalEndianness, err := strconv.ParseUint(string(signalEndiannessAndSigned[0]), 10, 1)
	if err != nil {
		return fmt.Errorf("signal line %d has invalid byte order: %s", i, string(signalEndiannessAndSigned[0]))
	}
	var signalSigned bool
	switch signalEndiannessAndSigned[1] {
	case '+':
		signalSigned = false
	case '-':
		signalSigned = true
	default:
		return fmt.Errorf("signal line %d has invalid signed: %s", i, string(signalEndiannessAndSigned[1]))
	}

	signal.Endianness = Endianness(signalEndianness)
	signal.Length = uint8(signalLength)
	signal.Signed = signalSigned
	signal.StartByte = uint8(signalStartByte)

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
