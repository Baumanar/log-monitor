package monitoring

import (
	"errors"
	"reflect"
	"testing"
)

func Test_parseLogLine(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		input   string
		want    *LogRecord
		wantErr error
	}{
		{"test1",
			"53.120.219.15 - paul [27/March/2020:12:10:41 +0100] \"GET /posts/r/a/view.html HTTP/1.0\" 403 5026",
			&LogRecord{
				remotehost: "53.120.219.15",
				rfc931:     "-",
				authuser:   "paul",
				date:       "[27/March/2020:12:10:41 +0100]",
				method:     "GET",
				section:    "/posts",
				protocol:   "HTTP/1.0",
				status:     "403",
				bytesCount: 5026,
			},
			nil,
		},

		{"test2",
			"141.146.202.67 - jill [27/March/2020:12:16:36 +0100] \"PUT /login/user?id=123 HTTP/1.0\" 200 1353",
			&LogRecord{
				remotehost: "141.146.202.67",
				rfc931:     "-",
				authuser:   "jill",
				date:       "[27/March/2020:12:16:36 +0100]",
				method:     "PUT",
				section:    "/login",
				protocol:   "HTTP/1.0",
				status:     "200",
				bytesCount: 1353,
			},
			nil,
		},
		// test with additional information
		{"test3",
			"141.146.202.67 1234 jill [27/March/2020:12:16:36 +0100] \"PUT /login/user?id=123 HTTP/1.0\" 200 1353 \"Mozilla/4.08 [en] (Win98; I ;Nav)\"",
			&LogRecord{
				remotehost: "141.146.202.67",
				rfc931:     "1234",
				authuser:   "jill",
				date:       "[27/March/2020:12:16:36 +0100]",
				method:     "PUT",
				section:    "/login",
				protocol:   "HTTP/1.0",
				status:     "200",
				bytesCount: 1353,
			},
			nil,
		},
		// test with bad formatting
		{"test4",
			"141.146.202.67 - jill [27/March/2020:12:16:36 +0100]",
			nil,
			errors.New("Invalid log format."),
		},

		{"test5",
			"[27/March/2020:12:16:36 +0100] 141.146.202.67 - jill ",
			nil,
			errors.New("Invalid log format."),
		},
		// test with empty string
		{"test6",
			"",
			nil,
			errors.New("Invalid log format."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLine(tt.input)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLogLine() \ngot = %v \nwant %v", got, tt.want)
			}
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("parseLogLine() err \ngot = %v \nwant %v", err, tt.wantErr)
			}

		})
	}
}
