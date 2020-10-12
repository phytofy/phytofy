// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for scheduling
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	indexStartDate = 0
	indexStopDate  = 1
	indexStartTime = 2
	indexStopTime  = 3
	indexLevels    = 4
)

type schdlTiming struct {
	Start uint32 `json:"start"`
	Stop  uint32 `json:"stop"`
}

type schdlLevels []float64

type schdlDetached struct {
	schdlTiming
	Levels schdlLevels `json:"levels"`
}

type schdlSerial uint32

type schdlSerials []schdlSerial

type schdlAttached struct {
	schdlDetached
	Serials schdlSerials `json:"serials"`
}

type schdlBlock struct {
	Begin    uint32
	End      uint32
	Schedule schdlDetached
}

type schdlAggregated map[schdlSerial][]schdlDetached

// Reads lines from a file
func schdlReadLinesFromFile(path string) ([]string, error) {
	file, fail := os.Open(path)
	if fail != nil {
		return nil, fmt.Errorf("Failed opening file %s: %s", path, fail)
	}
	defer file.Close()
	lines := make([]string, 0)
	reader := bufio.NewReader(file)
	for {
		var buffer bytes.Buffer
		var raw []byte
		var prefix bool
		for {
			raw, prefix, fail = reader.ReadLine()
			buffer.Write(raw)
			if !prefix || fail != nil {
				break
			}
		}
		lines = append(lines, strings.TrimSpace(buffer.String()))
		if fail == io.EOF {
			break
		} else if fail != nil {
			return nil, fmt.Errorf("Failed reading file %s: %s", path, fail)
		}
	}
	return lines, nil
}

// Reads schedules from CSV file and returns parsed
func schdlReadSchedulesFromFile(path string, channelCount int) ([]schdlAttached, error) {
	lines, fail := schdlReadLinesFromFile(path)
	if fail != nil {
		return nil, fail
	}
	schedules, fail := schdlParseLines(lines, channelCount)
	if fail != nil {
		return nil, fail
	}
	return schedules, nil
}

// Parses CSV lines into entries
func schdlParseLines(lines []string, channelCount int) ([]schdlAttached, error) {
	indexSerials := indexLevels + channelCount
	schedules := make([]schdlAttached, 0)
	for index, line := range lines {
		if len(line) == 0 && index == len(lines)-1 {
			continue
		}
		items := strings.Split(line, ",")
		count := len(items)
		if count < indexSerials {
			return nil, fmt.Errorf("Too few columns (%d) in line %d (%s)", count, index+1, line)
		}
		parsedLevels, fail := schdlParseLevels(items, channelCount)
		if fail != nil {
			return nil, fail
		}
		parsedSerials, fail := schdlParseSerials(items, indexSerials)
		if fail != nil {
			return nil, fail
		}
		parsedScheduling, fail := schdlParseScheduling(items)
		if fail != nil {
			return nil, fail
		}
		schedule := schdlAttached{schdlDetached{parsedScheduling, parsedLevels}, parsedSerials}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

// Converts human readable date and time to a timestamp
func schdlParseDate(atDate, atTime string) (uint32, error) {
	stampUtc := fmt.Sprintf("%sT%s.000Z", atDate, atTime)
	stamp, fail := time.Parse(time.RFC3339, stampUtc)
	if fail != nil {
		return 0, fmt.Errorf("Failed to parse time: %s", stampUtc)
	}
	return uint32(stamp.Unix()), nil
}

// Parses the scheduling portion of the schedule
func schdlParseScheduling(items []string) (schdlTiming, error) {
	start, fail := schdlParseDate(items[indexStartDate], items[indexStartTime])
	if fail != nil {
		return schdlTiming{0, 0}, fail
	}
	stop, fail := schdlParseDate(items[indexStopDate], items[indexStopTime])
	if fail != nil {
		return schdlTiming{0, 0}, fail
	}
	return schdlTiming{start, stop}, nil
}

// Parses the channel level portion of the schedule
func schdlParseLevels(items []string, channelCount int) (schdlLevels, error) {
	result := make(schdlLevels, channelCount)
	for i := 0; i < channelCount; i++ {
		level, fail := strconv.ParseInt(items[indexLevels+i], 10, 8)
		if fail != nil {
			return nil, fmt.Errorf("Cannot parse channel level: %s", items[indexLevels+i])
		}
		if level > 100 {
			return nil, fmt.Errorf("Invalid channel level (must be 0-100): %d", level)
		}
		result[i] = float64(level)
	}
	return result, nil
}

// Parses the serials portion of the schedule
func schdlParseSerials(items []string, indexSerials int) (schdlSerials, error) {
	serialsTextual := items[indexSerials:]
	serialsCount := len(serialsTextual)
	collection := make(map[schdlSerial]struct{})
	exists := struct{}{}
	for i := 0; i < serialsCount; i++ {
		serial, fail := strconv.ParseUint(serialsTextual[i], 10, 32)
		if fail != nil {
			return nil, fmt.Errorf("Cannot parse serial: %s", serialsTextual[i])
		}
		collection[schdlSerial(serial)] = exists
	}
	result := make(schdlSerials, 0, len(collection))
	for serial := range collection {
		result = append(result, serial)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}

// Splits the schedules by day
func schdlSplitSchedulesByDay(schedules []schdlAttached) []schdlAttached {
	daily := make([]schdlAttached, 0)
	for _, schedule := range schedules {
		start := schdlDropTime(schedule.Start)
		days, seconds := schdlCountDelta(schedule.Start, schedule.Stop)
		startTime := schdlDropDate(schedule.Start)
		stopTime := startTime + seconds
		for day := uint32(0); day < days+1; day++ {
			date := schdlShiftByDays(start, day)
			start := date + startTime
			stop := date + stopTime
			single := schdlAttached{schdlDetached{schdlTiming{start, stop}, schedule.Levels}, schedule.Serials}
			daily = append(daily, single)
		}
	}
	return daily
}

// Organizes schedules, checks for validity and possible overlap
func schdlCheckAllForValidity(schedules []schdlAttached) error {
	for _, schedule := range schedules {
		if fail := schdlCheckLevels(schedule.Levels); fail != nil {
			return fail
		}
		if fail := schdlCheckForValidity(schedule); fail != nil {
			return fail
		}
	}
	return nil
}

// Verifies if a schedule is valid
func schdlCheckForValidity(schedule schdlAttached) error {
	if schedule.Stop <= schedule.Start {
		return fmt.Errorf("Timespan invalid for schedule %+v", schedule)
	}
	return nil
}

// Checks aggregated schedules for possible overlap
func schdlCheckAllForOverlap(aggregated schdlAggregated) error {
	for serial := range aggregated {
		blocks := schdlExtractAllBlocks(aggregated[serial])
		sort.Slice(blocks, func(i, j int) bool { return blocks[i].Begin < blocks[j].Begin })
		i := 0
		for j := range blocks {
			if j == 0 {
				continue
			}
			if fail := schdlCheckForOverlap(blocks[i], blocks[j]); fail != nil {
				return fail
			}
			if blocks[i].End < blocks[j].End {
				i = j
			}
		}
	}
	return nil
}

// Checks two schedules for possible overlap
func schdlCheckForOverlap(blockX, blockY schdlBlock) error {
	if blockX.Begin < blockY.End && blockY.Begin < blockX.End {
		return fmt.Errorf("Schedules overlap (%+v and %+v)", blockX.Schedule, blockY.Schedule)
	}
	return nil
}

// Splits schedules into blocks considering daily repetition
func schdlExtractAllBlocks(schedules []schdlDetached) []schdlBlock {
	blocks := make([]schdlBlock, 0)
	for _, schedule := range schedules {
		blocks = append(blocks, schdlExtractBlocks(schedule)...)
	}
	return blocks
}

// Splits schedule into blocks considering daily repetition
func schdlExtractBlocks(schedule schdlDetached) []schdlBlock {
	blocks := make([]schdlBlock, 0)
	start := schdlDropTime(schedule.Start)
	days, seconds := schdlCountDelta(schedule.Start, schedule.Stop)
	startTime := schdlDropDate(schedule.Start)
	stopTime := startTime + seconds
	for day := uint32(0); day < days+1; day++ {
		date := schdlShiftByDays(start, day)
		begin := date + startTime
		end := date + stopTime
		blocks = append(blocks, schdlBlock{begin, end, schedule})
	}
	return blocks
}

// Drop time portion of timestamp
func schdlDropTime(timestamp uint32) uint32 {
	year, month, day := time.Unix(int64(timestamp), 0).UTC().Date()
	return uint32(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix())
}

// Drop date portion of timestamp
func schdlDropDate(timestamp uint32) uint32 {
	return timestamp - schdlDropTime(timestamp)
}

// Counts the number of days & seconds from start to stop
func schdlCountDelta(start, stop uint32) (uint32, uint32) {
	startAt := time.Unix(int64(start), 0).UTC()
	stopAt := time.Unix(int64(stop), 0).UTC()
	delta := stopAt.Sub(startAt)
	days := uint32(delta.Hours()) / 24
	seconds := uint32(delta.Seconds()) % (24 * 60 * 60)
	return days, seconds
}

// Shifts the timestamp by a given number of days
func schdlShiftByDays(timestamp uint32, days uint32) uint32 {
	delta := time.Hour * time.Duration(24*days)
	return uint32(time.Unix(int64(timestamp), 0).UTC().Add(delta).Unix())
}

// Aggregates a list of schedules by a serial
func schdlAggregateBySerial(schedules []schdlAttached) schdlAggregated {
	aggregated := make(schdlAggregated)
	for _, schedule := range schedules {
		inverted := schdlDetached{schedule.schdlTiming, schedule.Levels}
		for _, serial := range schedule.Serials {
			merged, present := aggregated[serial]
			if present {
				merged = append(merged, inverted)
			} else {
				merged = []schdlDetached{inverted}
			}
			aggregated[serial] = merged
		}
	}
	return aggregated
}

// Aggregates schedules per serial
func schdlAggregateSchedules(schedules []schdlAttached, splitSchedules bool) (schdlAggregated, error) {
	if fail := schdlCheckAllForValidity(schedules); fail != nil {
		return nil, fail
	}
	if splitSchedules {
		schedules = schdlSplitSchedulesByDay(schedules)
	}
	aggregated := schdlAggregateBySerial(schedules)
	if fail := schdlCheckAllForOverlap(aggregated); fail != nil {
		return nil, fail
	}
	return aggregated, nil
}

// Checks channels validity (individual range and total sum)
func schdlCheckLevels(levels schdlLevels) error {
	// Only % PWM allows to check the power reliably
	// but because the relation is almost linear
	// it is a good enough approximation
	total := uint16(0)
	for index := 0; index < len(levels); index++ {
		if levels[index] < 0 || levels[index] > 100 {
			return fmt.Errorf("Level at index %d out of bounds for levels - %v", index, levels)
		}
		total += uint16(levels[index])
	}
	if total > 300 {
		return fmt.Errorf("Cumulative irradiance exceeds 300%% for levels - %v", levels)
	}
	return nil
}
