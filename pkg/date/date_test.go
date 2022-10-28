package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEOD(t *testing.T) {
	assert := assert.New(t)

	middleOfDay := time.Date(2022, 10, 28, 10, 40, 0, 0, time.UTC)
	expectedEOD := time.Date(2022, 10, 28, 23, 59, 59, 0, time.UTC)

	actualEOD := EOD(middleOfDay, time.UTC)
	assert.Equal(expectedEOD, actualEOD)
}

func TestEOD_with_location(t *testing.T) {
	assert := assert.New(t)

	inputLoc, err := time.LoadLocation("America/New_York")
	assert.NoError(err)

	outputLoc, err := time.LoadLocation("America/Chicago")
	assert.NoError(err)

	middleOfDay := time.Date(2022, 10, 28, 10, 40, 0, 0, inputLoc)
	expectedEOD := time.Date(2022, 10, 28, 23, 59, 59, 0, outputLoc)

	actualEOD := EOD(middleOfDay, outputLoc)
	assert.Equal(expectedEOD, actualEOD)
}

func TestBOD(t *testing.T) {
	assert := assert.New(t)

	middleOfDay := time.Date(2022, 10, 28, 10, 40, 0, 0, time.UTC)
	expectedBOD := time.Date(2022, 10, 28, 0, 0, 0, 0, time.UTC)

	actualBOD := BOD(middleOfDay, time.UTC)
	assert.Equal(expectedBOD, actualBOD)
}

func TestBOD_with_location(t *testing.T) {
	assert := assert.New(t)

	inputLoc, err := time.LoadLocation("America/New_York")
	assert.NoError(err)

	outputLoc, err := time.LoadLocation("America/Chicago")
	assert.NoError(err)

	middleOfDay := time.Date(2022, 10, 28, 10, 40, 0, 0, inputLoc)
	expectedBOD := time.Date(2022, 10, 28, 0, 0, 0, 0, outputLoc)

	actualBOD := BOD(middleOfDay, outputLoc)
	assert.Equal(expectedBOD, actualBOD)
}

func TestDays(t *testing.T) {
	assert := assert.New(t)

	duration := time.Duration(24*3*time.Hour) + time.Duration(18*time.Hour)

	days := Days(duration)
	assert.Equal(3, days)
}

func TestWithinDuration_dates_within_duration(t *testing.T) {
	assert := assert.New(t)

	today := time.Now()
	yesterday := time.Now().Add(-1 * 24 * time.Hour)
	duration := time.Duration(48 * time.Hour)

	withinDuration := WithinDuration(today, yesterday, duration)
	assert.True(withinDuration)
}

func TestWithinDuration_dates_outside_duration(t *testing.T) {
	assert := assert.New(t)

	today := time.Now()
	yesterday := time.Now().Add(-1 * 24 * time.Hour)
	duration := time.Duration(1 * time.Second)

	withinDuration := WithinDuration(today, yesterday, duration)
	assert.False(withinDuration)
}

func TestParseAny(t *testing.T) {
	assert := assert.New(t)

	var approxDateLayouts = []string{"01/02/2006", "01/2006", "2006"}
	dateString := "10/2022"

	parsedDate, err := ParseAny(approxDateLayouts, dateString)
	assert.NoError(err)

	assert.Equal(time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), parsedDate)
}

func TestParseAny_error(t *testing.T) {
	assert := assert.New(t)

	var approxDateLayouts = []string{"01/02/2006", "01/2006", "2006"}
	dateString := "2022-10-01"

	_, err := ParseAny(approxDateLayouts, dateString)
	assert.ErrorIs(err, ErrNoLayoutMatched)
}
