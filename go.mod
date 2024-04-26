module github.com/skillian/textutil

go 1.19

require (
	github.com/atotto/clipboard v0.1.4
	github.com/skillian/argparse v0.0.0-20220329092431-5b00f5285c1e
	github.com/skillian/errors v0.0.0-20220617152528-c0cc06767f86
)

require (
	github.com/skillian/logging v0.0.0-20210406222847-057884e2cfcc // indirect
	github.com/skillian/textwrap v0.0.0-20190707153458-15c7ee8d44ed // indirect
)

replace "github.com/skillian/argparse" => "../argparse"