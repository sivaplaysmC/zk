package zk

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestParseDefaultConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(""), "")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
		Editor: opt.NullString,
		DirConfig: DirConfig{
			FilenameTemplate: "{{id}}",
			Extension:        "md",
			BodyTemplatePath: opt.NullString,
			IDOptions: IDOptions{
				Length:  5,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			},
			DefaultTitle: "Untitled",
			Lang:         "en",
			Extra:        make(map[string]string),
		},
		Dirs:    make(map[string]DirConfig),
		Aliases: make(map[string]string),
	})
}

func TestParseInvalidConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(`;`), "")

	assert.NotNil(t, err)
	assert.Nil(t, conf)
}

func TestParseComplete(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		# Comment
		editor = "vim"
		pager = "less"
		filename = "{{id}}.note"
		extension = "txt"
		template = "default.note"
		language = "fr"
		default-title = "Sans titre"

		[id]
		charset = "alphanum"
		length = 4
		case = "lower"

		[extra]
		hello = "world"
		salut = "le monde"

		[alias]
		ls = "zk list $@"
		ed = "zk edit $@"

		[dir.log]
		filename = "{{date}}.md"
		extension = "note"
		template = "log.md"
		language = "de"
		default-title = "Ohne Titel"

		[dir.log.id]
		charset = "letters"
		length = 8
		case = "mixed"
		
		[dir.log.extra]
		log-ext = "value"

		[dir.ref]
		filename = "{{slug title}}.md"
	`), "")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
		DirConfig: DirConfig{
			FilenameTemplate: "{{id}}.note",
			Extension:        "txt",
			BodyTemplatePath: opt.NewString("default.note"),
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			},
			Lang:         "fr",
			DefaultTitle: "Sans titre",
			Extra: map[string]string{
				"hello": "world",
				"salut": "le monde",
			},
		},
		Dirs: map[string]DirConfig{
			"log": {
				FilenameTemplate: "{{date}}.md",
				Extension:        "note",
				BodyTemplatePath: opt.NewString("log.md"),
				IDOptions: IDOptions{
					Length:  8,
					Charset: CharsetLetters,
					Case:    CaseMixed,
				},
				Lang:         "de",
				DefaultTitle: "Ohne Titel",
				Extra: map[string]string{
					"hello":   "world",
					"salut":   "le monde",
					"log-ext": "value",
				},
			},
			"ref": {
				FilenameTemplate: "{{slug title}}.md",
				Extension:        "txt",
				BodyTemplatePath: opt.NewString("default.note"),
				IDOptions: IDOptions{
					Length:  4,
					Charset: CharsetAlphanum,
					Case:    CaseLower,
				},
				Lang:         "fr",
				DefaultTitle: "Sans titre",
				Extra: map[string]string{
					"hello": "world",
					"salut": "le monde",
				},
			},
		},
		Editor: opt.NewString("vim"),
		Pager:  opt.NewString("less"),
		Aliases: map[string]string{
			"ls": "zk list $@",
			"ed": "zk edit $@",
		},
	})
}

func TestParseMergesDirConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		filename = "root-filename"
		extension = "txt"
		template = "root-template"
		language = "fr"
		default-title = "Sans titre"

		[id]
		charset = "letters"
		length = 42
		case = "upper"

		[extra]
		hello = "world"
		salut = "le monde"

		[dir.log]
		filename = "log-filename"
		template = "log-template"

		[dir.log.id]
		charset = "numbers"
		length = 8
		case = "mixed"

		[dir.log.extra]
		hello = "override"
		log-ext = "value"

		[dir.inherited]
	`), "")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
		DirConfig: DirConfig{
			FilenameTemplate: "root-filename",
			Extension:        "txt",
			BodyTemplatePath: opt.NewString("root-template"),
			IDOptions: IDOptions{
				Length:  42,
				Charset: CharsetLetters,
				Case:    CaseUpper,
			},
			Lang:         "fr",
			DefaultTitle: "Sans titre",
			Extra: map[string]string{
				"hello": "world",
				"salut": "le monde",
			},
		},
		Dirs: map[string]DirConfig{
			"log": {
				FilenameTemplate: "log-filename",
				Extension:        "txt",
				BodyTemplatePath: opt.NewString("log-template"),
				IDOptions: IDOptions{
					Length:  8,
					Charset: CharsetNumbers,
					Case:    CaseMixed,
				},
				Lang:         "fr",
				DefaultTitle: "Sans titre",
				Extra: map[string]string{
					"hello":   "override",
					"salut":   "le monde",
					"log-ext": "value",
				},
			},
			"inherited": {
				FilenameTemplate: "root-filename",
				Extension:        "txt",
				BodyTemplatePath: opt.NewString("root-template"),
				IDOptions: IDOptions{
					Length:  42,
					Charset: CharsetLetters,
					Case:    CaseUpper,
				},
				Lang:         "fr",
				DefaultTitle: "Sans titre",
				Extra: map[string]string{
					"hello": "world",
					"salut": "le monde",
				},
			},
		},
		Aliases: make(map[string]string),
	})
}

func TestParseIDCharset(t *testing.T) {
	test := func(charset string, expected Charset) {
		toml := fmt.Sprintf(`
			[id]
			charset = "%v"
		`, charset)
		conf, err := ParseConfig([]byte(toml), "")
		assert.Nil(t, err)
		if !cmp.Equal(conf.IDOptions.Charset, expected) {
			t.Errorf("Didn't parse ID charset `%v` as expected", charset)
		}
	}

	test("alphanum", CharsetAlphanum)
	test("hex", CharsetHex)
	test("letters", CharsetLetters)
	test("numbers", CharsetNumbers)
	test("HEX", []rune("HEX")) // case sensitive
	test("custom", []rune("custom"))
}

func TestParseIDCase(t *testing.T) {
	test := func(letterCase string, expected Case) {
		toml := fmt.Sprintf(`
			[id]
			case = "%v"
		`, letterCase)
		conf, err := ParseConfig([]byte(toml), "")
		assert.Nil(t, err)
		if !cmp.Equal(conf.IDOptions.Case, expected) {
			t.Errorf("Didn't parse ID case `%v` as expected", letterCase)
		}
	}

	test("lower", CaseLower)
	test("upper", CaseUpper)
	test("mixed", CaseMixed)
	test("unknown", CaseLower)
}

func TestParseResolvesTemplatePaths(t *testing.T) {
	test := func(template string, expected string) {
		toml := fmt.Sprintf(`template = "%v"`, template)
		conf, err := ParseConfig([]byte(toml), "/test/.zk/templates")
		assert.Nil(t, err)
		if !cmp.Equal(conf.BodyTemplatePath, opt.NewString(expected)) {
			t.Errorf("Didn't resolve template `%v` as expected: %v", template, conf.BodyTemplatePath)
		}
	}

	test("template.tpl", "/test/.zk/templates/template.tpl")
	test("/abs/template.tpl", "/abs/template.tpl")
}

func TestDirConfigClone(t *testing.T) {
	original := DirConfig{
		FilenameTemplate: "{{id}}.note",
		Extension:        "md",
		BodyTemplatePath: opt.NewString("default.note"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetAlphanum,
			Case:    CaseLower,
		},
		Lang:         "fr",
		DefaultTitle: "Sans titre",
		Extra: map[string]string{
			"hello": "world",
		},
	}

	clone := original.Clone()
	// Check that the clone is equivalent
	assert.Equal(t, clone, original)

	clone.FilenameTemplate = "modified"
	clone.Extension = "txt"
	clone.BodyTemplatePath = opt.NewString("modified")
	clone.IDOptions.Length = 41
	clone.IDOptions.Charset = CharsetNumbers
	clone.IDOptions.Case = CaseUpper
	clone.Lang = "de"
	clone.DefaultTitle = "Ohne Titel"
	clone.Extra["test"] = "modified"

	// Check that we didn't modify the original
	assert.Equal(t, original, DirConfig{
		FilenameTemplate: "{{id}}.note",
		Extension:        "md",
		BodyTemplatePath: opt.NewString("default.note"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetAlphanum,
			Case:    CaseLower,
		},
		Lang:         "fr",
		DefaultTitle: "Sans titre",
		Extra: map[string]string{
			"hello": "world",
		},
	})
}

func TestDirConfigOverride(t *testing.T) {
	sut := DirConfig{
		FilenameTemplate: "filename",
		BodyTemplatePath: opt.NewString("body.tpl"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetLetters,
			Case:    CaseUpper,
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	}

	// Empty overrides
	sut.Override(ConfigOverrides{})
	assert.Equal(t, sut, DirConfig{
		FilenameTemplate: "filename",
		BodyTemplatePath: opt.NewString("body.tpl"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetLetters,
			Case:    CaseUpper,
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	})

	// Some overrides
	sut.Override(ConfigOverrides{
		BodyTemplatePath: opt.NewString("overriden-template"),
		Extra: map[string]string{
			"hello":      "overriden",
			"additional": "value",
		},
	})
	assert.Equal(t, sut, DirConfig{
		FilenameTemplate: "filename",
		BodyTemplatePath: opt.NewString("overriden-template"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetLetters,
			Case:    CaseUpper,
		},
		Extra: map[string]string{
			"hello":      "overriden",
			"salut":      "le monde",
			"additional": "value",
		},
	})
}
