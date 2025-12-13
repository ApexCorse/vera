package vera

import (
	"strconv"
	"strings"
)

type Signal struct {
	Name     string
	StartBit uint8
	Length   uint8
	Endianness
	Signed    bool
	Factor    float32
	Offset    float32
	Min       float32
	Max       float32
	Unit      string
	Receivers []Node
	Topic     string

	lineNumber int
}

func (s *Signal) Validate() error {
	if s.StartBit >= 64 {
		return errorAtLine(s.lineNumber, "signal start bit must be a number between 0 and 63")
	}
	if s.Length > 64 {
		return errorAtLine(s.lineNumber, "signal length must be a number between 1 and 64")
	}
	if s.Factor == 0 {
		return errorAtLine(s.lineNumber, "signal factor cannot be zero")
	}

	return nil
}

func NewSignalFromLine(message *Message, line string, lineNumber int) (*Signal, error) {
	signal := &Signal{
		lineNumber: lineNumber,
	}

	line = strings.TrimFunc(line, func(r rune) bool {
		return r == ' ' || r == '\t'
	})

	if !strings.HasPrefix(line, "SG_") {
		return nil, errorAtLine(signal.lineNumber, "signal line doesn't start with 'SG_'")
	}

	lineParts := strings.Fields(line)
	if err := signal.checkLineStructure(lineParts); err != nil {
		return nil, err
	}

	signal.Name = lineParts[1]

	if err := signal.parseBitInfo(message, lineParts[3]); err != nil {
		return nil, err
	}

	if err := signal.parseFactorOffset(lineParts[4]); err != nil {
		return nil, err
	}

	if err := signal.parseMinMax(lineParts[5]); err != nil {
		return nil, err
	}

	if err := signal.parseUnit(lineParts[6]); err != nil {
		return nil, err
	}

	signal.parseReceivers(lineParts[7])

	return signal, nil
}

func (s *Signal) checkLineStructure(lineParts []string) error {
	if len(lineParts) < 7 {
		return errorAtLine(s.lineNumber, `signal line is not well structured, must adhere to:
SG_ <SignalName> : <StartBit>|<Length>@<BitOrder><Signed> (<Factor>,<Offset>) [<Min>,<Max>] "<Unit>" <...Receivers>`)
	}

	if lineParts[2] != ":" {
		return errorAtLine(s.lineNumber, "signal line has not a ':' between <SignalName> and <StartBit>")
	}

	return nil
}

func (s *Signal) parseBitInfo(message *Message, signalBitInfo string) error {
	signalBitFirstSplit := strings.Split(signalBitInfo, "@")
	if len(signalBitFirstSplit) != 2 {
		return errorAtLine(s.lineNumber, "signal line has invalid bit info: %s", signalBitInfo)
	}

	signalStartBitAndLength := strings.Split(signalBitFirstSplit[0], "|")
	if len(signalStartBitAndLength) != 2 {
		return errorAtLine(s.lineNumber, "signal line has invalid start bit and length: %s", signalBitFirstSplit[0])
	}
	signalStartBit, err := strconv.Atoi(signalStartBitAndLength[0])
	if err != nil {
		return errorAtLine(s.lineNumber, "signal line has invalid start bit: %s", signalStartBitAndLength[0])
	}
	signalLength, err := strconv.Atoi(signalStartBitAndLength[1])
	if err != nil {
		return errorAtLine(s.lineNumber, "signal line has invalid length: %s", signalStartBitAndLength[1])
	}

	signalOtherInfo := signalBitFirstSplit[1]
	if len(signalOtherInfo) < 2 {
		return errorAtLine(s.lineNumber, "signal line has invalid bit order and signed: %s", signalOtherInfo)
	}

	signalEndianness, err := strconv.ParseUint(string(signalOtherInfo[0]), 10, 1)
	if err != nil {
		return errorAtLine(s.lineNumber, "signal line has invalid bit order: %s", string(signalOtherInfo[0]))
	}
	var signalSigned bool
	switch signalOtherInfo[1] {
	case '+':
		signalSigned = false
	case '-':
		signalSigned = true
	default:
		return errorAtLine(s.lineNumber, "signal line has invalid signed: %s", string(signalOtherInfo[1]))
	}

	s.Endianness = Endianness(signalEndianness)
	s.Length = uint8(signalLength)
	s.Signed = signalSigned
	s.StartBit = uint8(signalStartBit)
	message.signalsTotalLength += uint8(signalLength)
	message.signalsBitPositions[signalStartBit] = len(message.Signals)
	message.signalsBitPositions[signalStartBit+signalLength-1] = len(message.Signals)

	return nil
}

func (s *Signal) parseFactorOffset(signalFactorOffset string) error {
	if !strings.HasPrefix(signalFactorOffset, "(") || !strings.HasSuffix(signalFactorOffset, ")") {
		return errorAtLine(s.lineNumber, "signal line has invalid factor and offset: %s", signalFactorOffset)
	}

	signalFactorOffsetParts := strings.Split(signalFactorOffset[1:len(signalFactorOffset)-1], ",")
	if len(signalFactorOffsetParts) != 2 {
		return errorAtLine(s.lineNumber, "signal line has invalid factor and offset: %s", signalFactorOffset)
	}

	signalFactor, err := strconv.ParseFloat(signalFactorOffsetParts[0], 32)
	if err != nil {
		return errorAtLine(s.lineNumber, "signal line has invalid factor: %s", string(signalFactorOffsetParts[0]))
	}

	signalOffset, err := strconv.ParseFloat(signalFactorOffsetParts[1], 32)
	if err != nil {
		return errorAtLine(s.lineNumber, "signal line has invalid offset: %s", string(signalFactorOffsetParts[1]))
	}

	s.Factor = float32(signalFactor)
	s.Offset = float32(signalOffset)

	return nil
}

func (s *Signal) parseMinMax(signalMinMax string) error {
	if !strings.HasPrefix(signalMinMax, "[") || !strings.HasSuffix(signalMinMax, "]") {
		return errorAtLine(s.lineNumber, "signal line has invalid min/max: %s", signalMinMax)
	}

	signalMinMaxParts := strings.Split(signalMinMax[1:len(signalMinMax)-1], "|")
	if len(signalMinMaxParts) != 2 {
		return errorAtLine(s.lineNumber, "signal line has invalid min/max: %s", signalMinMax)
	}

	signalMin, err := strconv.ParseFloat(signalMinMaxParts[0], 32)
	if err != nil {
		return errorAtLine(s.lineNumber, "signal line has invalid min: %s", string(signalMinMaxParts[0]))
	}
	signalMax, err := strconv.ParseFloat(signalMinMaxParts[1], 32)
	if err != nil {
		return errorAtLine(s.lineNumber, "signal line has invalid max: %s", string(signalMinMaxParts[1]))
	}

	s.Min = float32(signalMin)
	s.Max = float32(signalMax)

	return nil
}

func (s *Signal) parseUnit(signalUnitStr string) error {
	if !strings.HasPrefix(signalUnitStr, "\"") || !strings.HasSuffix(signalUnitStr, "\"") {
		return errorAtLine(s.lineNumber, "signal line has invalid unit: %s", signalUnitStr)
	}

	s.Unit = signalUnitStr[1 : len(signalUnitStr)-1]

	return nil
}

func (s *Signal) parseReceivers(signalReceivers string) {
	receiversStr := strings.Split(signalReceivers, ",")

	for _, r := range receiversStr {
		s.Receivers = append(s.Receivers, Node(r))
	}
}
