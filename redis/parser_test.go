package redis_test

import (
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
)

func TestParser(t *testing.T) {
	t.Run("ECHO request", func(t *testing.T) {
		testCases := []struct {
			name    string
			request string
			typ     redis.Type
		}{
			{
				name:    "ECHO hey",
				request: "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
			},
			{
				name:    "echo hey",
				request: "*2\r\n$4\r\necho\r\n$3\r\nhey\r\n",
			},
			{
				name:    "EcHo hey",
				request: "*2\r\n$4\r\nEcHo\r\n$3\r\nhey\r\n",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				reader := strings.NewReader(testCase.request)

				command, err := redis.Parser{}.Parse(reader)

				if err != nil {
					t.Errorf("err: expected: nil; got: %v", err)
				}

				// TODO: finish tests and migrating to Command2
				_ = command
				//if _, ok := command.(redis.EchoCommand2); !ok {
				//	t.Errorf(`command.Type() expected to be redis.Echo but was "%v"`, command.Type())
				//}
			})
		}
	})

	//		t.Run("and executing it", func(t *testing.T) {
	//			t.Run("then it writes the ECHO request argument", func(t *testing.T) {
	//				echoRequest := "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"
	//				reader := strings.NewReader(echoRequest)
	//				var writer strings.Builder
	//
	//				command, _ := redis.Parser{}.Parse(reader)
	//				err := command.Run(&writer)
	//
	//				if err != nil {
	//					t.Errorf("err expected to be nil but was %v", err)
	//				}
	//				if writer.String() != "$3\r\nhey\r\n" {
	//					t.Errorf(`command.Run() expected to write "$3\r\nhey\r\n" but was "%v"`, writer.String())
	//				}
	//			})
	//		})
	//	})
	//})
	//
	//t.Run("given a PING request", func(t *testing.T) {
	//	t.Run("when parsing it", func(t *testing.T) {
	//		t.Run("then it returns the redis.Ping command type", func(t *testing.T) {
	//			testCases := []struct {
	//				name    string
	//				request string
	//				typ     redis.Type
	//			}{
	//				{
	//					name:    "PING",
	//					request: "*1\r\n$4\r\nPING\r\n",
	//					typ:     redis.Ping,
	//				},
	//				{
	//					name:    "ping",
	//					request: "*1\r\n$4\r\nping\r\n",
	//					typ:     redis.Ping,
	//				},
	//				{
	//					name:    "PiNg",
	//					request: "*1\r\n$4\r\nPiNg\r\n",
	//					typ:     redis.Ping,
	//				},
	//			}
	//
	//			for _, testCase := range testCases {
	//				t.Run(testCase.name, func(t *testing.T) {
	//					reader := strings.NewReader(testCase.request)
	//
	//					command, err := redis.Parser{}.Parse(reader)
	//
	//					if err != nil {
	//						t.Errorf("err expected to be nil but was %v", err)
	//					}
	//					if command.Type() != redis.Ping {
	//						t.Errorf(`command.Type() expected to be redis.Ping but was "%v"`, command.Type())
	//					}
	//				})
	//			}
	//		})
	//
	//		t.Run("and executing it", func(t *testing.T) {
	//			t.Run("then it writes PONG", func(t *testing.T) {
	//				pingRequest := "*1\r\n$4\r\nPING\r\n"
	//				reader := strings.NewReader(pingRequest)
	//				var writer strings.Builder
	//
	//				command, _ := redis.Parser{}.Parse(reader)
	//				err := command.Run(&writer)
	//
	//				if err != nil {
	//					t.Errorf("err expected to be nil but was %v", err)
	//				}
	//				if writer.String() != "+PONG\r\n" {
	//					t.Errorf(`command.Run() expected to write "+PONG\r\n" but was "%v"`, writer.String())
	//				}
	//			})
	//		})
	//	})
	//})
}
