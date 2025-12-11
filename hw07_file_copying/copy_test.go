package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var text = `Bury Tomorrow - Glasswalk
I've seen the light at the end of the walk
These steps lead through the dark
Here now I stand, this chance in hand
To be the only one to hold on to, follow you
In the fear I am shrouded
Consumed, it's suffocating me
It's only you
(It's only you)
That can get us through
Is this the only chance to start new?
At the end of it all, once the roof begins to fall
Will we then know what we were meant to be?
Or are we fated to pace the hall?
Broken glass at our feet, we are forced to our knees
I've seen the light at the end of the walk
These steps lead through the dark
Here now I stand, this chance in hand
To be the only one to hold on to, follow you
It's never ending, there's no escaping it
Breathe in, breathe out, and make the memory
Its never ending, there's no avoiding it
Breathe in, breath out, and make the memory
Im not sure if I'm willing or able
To reach the door that so many pray for
A constant battle with the fear of falling
My eyes are open to a world that's calling
I've seen the light at the end of the walk
These steps lead through the dark
Here now I stand, this chance in hand
To be the only one to hold on to, follow you
I know it feels like hell
To gain the strength to walk again
Here now I stand, this chance in hand
To be the only one to hold on to, follow you
Take a step, make the move, ask yourself
What the fuck have you got to lose
What the fuck have you got to lose
I've seen the light at the end of the walk
These steps lead through the dark
Here now I stand, this chance in hand
To be the only one to hold on to, follow you
I know it feels like hell
To gain the strength to walk again
Here now I stand, this chance in hand
To be the only one to hold on to, follow you`

// Size of text: 1727

func TestCopy(t *testing.T) {
	srcFile := "/tmp/testSrc"
	destFile := "/tmp/testDest"
	testCases := []struct {
		offset   int64
		limit    int64
		expected []byte
	}{
		{
			offset: 0, limit: 25,
			expected: []byte(`Bury Tomorrow - Glasswalk`),
		},
		{
			offset: 26, limit: 42,
			expected: []byte(`I've seen the light at the end of the walk`),
		},
		{
			offset: 1645, limit: 0,
			expected: []byte(`Here now I stand, this chance in hand
To be the only one to hold on to, follow you`),
		},
		{
			offset: 1645, limit: 82,
			expected: []byte(`Here now I stand, this chance in hand
To be the only one to hold on to, follow you`),
		},
		{
			offset: 1645, limit: 10_000,
			expected: []byte(`Here now I stand, this chance in hand
To be the only one to hold on to, follow you`),
		},
		{
			offset: 0, limit: 0,
			expected: []byte(text),
		},
		{
			offset: 0, limit: 10_000,
			expected: []byte(text),
		},
		{
			offset: 1727, limit: 0,
			expected: []byte{}, // будет создан файл, но пустой
		},
		{
			offset: 1727, limit: 1,
			expected: []byte{}, // будет создан файл, но пустой
		},
		{
			offset: 1727, limit: 10_000,
			expected: []byte{}, // будет создан файл, но пустой
		},
	}
	err := os.WriteFile(srcFile, []byte(text), 0o644)
	require.NoError(t, err)
	for _, tc := range testCases {
		t.Run("copy_testing", func(t *testing.T) {
			err = Copy(srcFile, destFile, tc.offset, tc.limit)
			require.NoError(t, err)
			data, err := os.ReadFile(destFile)
			require.NoError(t, err)
			require.Equal(t, tc.expected, data)
		})
	}
}

func TestErrors(t *testing.T) {
	srcFile := "/tmp/testSrcErr"
	destFile := "/tmp/testDestErr"
	testCases := []struct {
		srcFile  string
		offset   int64
		limit    int64
		expected error
	}{
		{
			srcFile: "/tmp/testSrcErr",
			offset:  1728, limit: 0,
			expected: ErrOffsetExceedsFileSize,
		},
		{
			srcFile: "/dev/null",
			offset:  0, limit: 0,
			expected: ErrUnsupportedFile,
		},
		{
			srcFile: "/tmp",
			offset:  0, limit: 0,
			expected: ErrUnsupportedFile,
		},
	}
	err := os.WriteFile(srcFile, []byte(text), 0o644)
	require.NoError(t, err)
	for _, tc := range testCases {
		t.Run("errors_testing", func(t *testing.T) {
			err = Copy(tc.srcFile, destFile, tc.offset, tc.limit)
			require.Error(t, err)
			require.Equal(t, tc.expected, err)
		})
	}
}
