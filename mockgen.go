//go:build gomock || generate

package moqtransport

//go:generate sh -c "go run go.uber.org/mock/mockgen -build_flags=\"-tags=gomock\" -package moqtransport -self_package github.com/danielpfeifer02/priority-moqtransport -destination mock_stream_test.go github.com/danielpfeifer02/priority-moqtransport Stream"
type Stream = stream

//go:generate sh -c "go run go.uber.org/mock/mockgen -build_flags=\"-tags=gomock\" -package moqtransport -self_package github.com/danielpfeifer02/priority-moqtransport -destination mock_receive_stream_test.go github.com/danielpfeifer02/priority-moqtransport ReceiveStream"
type ReceiveStream = receiveStream

//go:generate sh -c "go run go.uber.org/mock/mockgen -build_flags=\"-tags=gomock\" -package moqtransport -self_package github.com/danielpfeifer02/priority-moqtransport -destination mock_send_stream_test.go github.com/danielpfeifer02/priority-moqtransport SendStream"
type SendStream = sendStream

//go:generate sh -c "go run go.uber.org/mock/mockgen -build_flags=\"-tags=gomock\" -package moqtransport -self_package github.com/danielpfeifer02/priority-moqtransport -destination mock_connection_test.go github.com/danielpfeifer02/priority-moqtransport Connection"
type Connection = connection

//go:generate sh -c "go run go.uber.org/mock/mockgen -build_flags=\"-tags=gomock\" -package moqtransport -self_package github.com/danielpfeifer02/priority-moqtransport -destination mock_parser_test.go github.com/danielpfeifer02/priority-moqtransport Parser"
type Parser = parser

//go:generate sh -c "go run go.uber.org/mock/mockgen -build_flags=\"-tags=gomock\" -package moqtransport -self_package github.com/danielpfeifer02/priority-moqtransport -destination mock_parser_factory_test.go github.com/danielpfeifer02/priority-moqtransport ParserFactory"
type ParserFactory = parserFactory

//go:generate sh -c "go run go.uber.org/mock/mockgen -build_flags=\"-tags=gomock\" -package moqtransport -self_package github.com/danielpfeifer02/priority-moqtransport -destination mock_control_stream_handler_test.go github.com/danielpfeifer02/priority-moqtransport ControlStreamHandler"
type ControlStreamHandler = controlStreamHandler
