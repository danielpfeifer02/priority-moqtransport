module github.com/danielpfeifer02/priority-moqtransport

go 1.22.0

require (
	github.com/danielpfeifer02/quic-go-prio-packs v0.41.0-28
	github.com/stretchr/testify v1.8.4
	go.uber.org/goleak v1.3.0
	go.uber.org/mock v0.3.0
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63
)

replace github.com/danielpfeifer02/quic-go-prio-packs v0.41.0-28 => ../../quic-go-prio-packs

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/tools v0.12.1-0.20230815132531-74c255bcf846 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
