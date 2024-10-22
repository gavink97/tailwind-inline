package tailwindinline

import (
	"strings"
	"testing"
)

var styles = []byte(`
    .bg-zinc-50 {
    --tw-bg-opacity: 1;
    background-color: rgb(250 250 250 / var(--tw-bg-opacity));
    }

    .text-2xl {
    font-size: 1.5rem;
    line-height: 2rem;
    }

    .font-bold {
    font-weight: 700;
    }

    .text-black {
    --tw-text-opacity: 1;
    color: rgb(0 0 0 / var(--tw-text-opacity));
    }

    .text-sky-200 {
    --tw-text-opacity: 1;
    color: rgb(186 230 253 / var(--tw-text-opacity));
    }

    @media (min-width: 640px) {
    .sm\:text-black {
    --tw-text-opacity: 1;
    color: rgb(0 0 0 / var(--tw-text-opacity));
    }
    }

    @media (min-width: 1024px) {
    .lg\:text-2xl {
    font-size: 1.5rem;
    line-height: 2rem;
    }
    }

@media (max-width: 460px) {
  .max-\[460px\]\:block {
    display: block;
  }

  .max-\[460px\]\:w-full {
    width: 100%;
  }
}

@media (min-width: 1024px) {
  @media (min-width: 640px) {
    .lg\:sm\:bg-zinc-50 {
      --tw-bg-opacity: 1;
      background-color: rgb(250 250 250 / var(--tw-bg-opacity));
    }
  }
}
    `)

func TestGetClasses(t *testing.T) {
	tpl := `<h1 class="text-2xl p-4 font-bold">Testing</h1>`
	expected := []string{"text-2xl", "p-4", "font-bold"}

	result := getClasses(tpl)

	for index := range result {
		if result[index] != expected[index] {
			t.Errorf("Incorrect Result, result: %s expected: %s", result[index], expected[index])
		}
	}
}

func TestAssignField(t *testing.T) {
	classes := []string{"lg:p-8", "max-w-md", "@xs:p-8"}
	expected := []field{{Type: "media", Class: "lg:p-8", Prefix: "lg:"},
		{Type: "inline", Class: "max-w-md", Prefix: ""},
		{Type: "container", Class: "@xs:p-8", Prefix: "@xs:"}}

	for index, value := range classes {
		result := assignField(value)

		if result != expected[index] {
			t.Errorf("Incorrect Result, result: %s expected: %s", result, expected[index])
		}
	}
}

func TestGetInlineStyles(t *testing.T) {
	field := field{Type: "inline", Class: "text-black"}

	expected := `--tw-text-opacity: 1;
    color: rgb(0 0 0 / var(--tw-text-opacity));`

	result := getInlineStyles(styles, field)

	expected = strings.TrimSpace(expected)
	result = strings.TrimSpace(result)

	if !strings.EqualFold(result, expected) {
		t.Errorf("Incorrect Result, result: %s expected: %s", result, expected)
	}
}

func TestGetMediaQueries(t *testing.T) {
	field := field{Type: "media", Class: "max-[460px]:block", Prefix: "max-[460px]:"}

	expected := `@media (max-width: 460px) {
  .max-\[460px\]\:block {
    display: block;
  }

  .max-\[460px\]\:w-full {
    width: 100%;
  }
}`

	result := getMediaQueries(styles, field)

	expected = strings.TrimSpace(expected)
	result = strings.TrimSpace(result)

	if !strings.EqualFold(result, expected) {
		t.Errorf("Incorrect Result, result: %s expected: %s", result, expected)
	}
}

func TestValidateClassByFirstIndex(t *testing.T) {
	tpl := []string{`<h1 class="text-2xl p-4 font-bold sm:text-black">Hello Gavin</h1>`,
		`<a href="https://github.com/gavink97" class="sm:text-black lg:text-2xl font-bold">Github</a>`}

	field := []field{{Type: "inline", Class: "text-black"}, {Type: "media", Class: "lg:text-2xl"}}

	expected := []int{-1, 59}

	for index, value := range tpl {
		result := validateClassByFirstIndex(value, field[index])

		if result != expected[index] {
			t.Errorf("Incorrect Result, result: %d expected: %d", result, expected[index])
		}
	}
}

func TestReplaceWithInlineCSS(t *testing.T) {
	tpl := `<h1 class="text-2xl p-4 text-black">Hello</h1>`

	css := "--tw-text-opacity: 1;" + "\ncolor: rgb(0 0 0 / var(--tw-text-opacity));"

	field := field{Type: "inline", Class: "text-black"}

	expected := `style="--tw-text-opacity: 1;color: rgb(0 0 0 / var(--tw-text-opacity));"`

	result := replaceWithInlineCSS(tpl, css, field)

	if !strings.Contains(result, expected) {
		t.Errorf("Incorrect Result, result: %s expected: %s", result, expected)
	}
}

func TestWriteMediaQueries(t *testing.T) {
	tpl := "<style></style>\n" + `<h1 class="lg:text-2xl p-8 sm:text-black">Hello</h1>`

	css := []string{`@media (min-width: 640px) {
    .sm\:text-black {
    --tw-text-opacity: 1;
    color: rgb(0 0 0 / var(--tw-text-opacity));
    }
    }`, `@media (min-width: 1024px) {
    .lg\:text-2xl {
    font-size: 1.5rem;
    line-height: 2rem;
    }
    }`}

	expected := "<style>" + `
@media (min-width: 640px) {
    .sm\:text-black {
    --tw-text-opacity: 1;
    color: rgb(0 0 0 / var(--tw-text-opacity));
    }
    }
@media (min-width: 1024px) {
    .lg\:text-2xl {
    font-size: 1.5rem;
    line-height: 2rem;
    }
    }` + "\n</style>\n" + `<h1 class="lg:text-2xl p-8 sm:text-black">Hello</h1>`

	result := writeMediaQueries(tpl, css)

	if !strings.Contains(result, expected) {
		t.Errorf("Incorrect Result, result: %s expected: %s", result, expected)
	}
}

func TestRemoveOldClassTags(t *testing.T) {
	tpl := []string{`<body><h1 class=" " style="font-size: 1.5rem;line-height: ` +
		`2rem;font-weight: 700;">Hello </h1><p class=" " style="--tw-text-opacity: ` +
		`1;color: rgb(239 68 68 / var(--tw-text-opacity));font-size: ` +
		`1.25rem;line-height: 1.75rem;">This is a test is it working</p></body>`,

		`<body><h1 class=" " style="font-size: 1.5rem;line-height: ` +
			`2rem;">Hello </h1><p class=" md:font-bold lg:font-bold" style="--tw-text-opacity: 1;color: ` +
			`rgb(239 68 68 / var(--tw-text-opacity));font-size: 1.25rem;line-height: ` +
			`1.75rem;">This is a test is it working</p></body>`,

		`<body><h1 class=" md:font-bold" style="font-size: 1.5rem;line-height: ` +
			`2rem;">Hello </h1><p class=" " style="--tw-text-opacity: 1;color: ` +
			`rgb(239 68 68 / var(--tw-text-opacity));font-size: 1.25rem;line-height: ` +
			`1.75rem;">This is a test is it working</p></body>`}

	expected := []string{`<body><h1 style="font-size: 1.5rem;line-height: ` +
		`2rem;font-weight: 700;">Hello </h1><p style="--tw-text-opacity: ` +
		`1;color: rgb(239 68 68 / var(--tw-text-opacity));font-size: ` +
		`1.25rem;line-height: 1.75rem;">This is a test is it working</p></body>`,

		`<body><h1 style="font-size: 1.5rem;line-height: ` +
			`2rem;">Hello </h1><p class="md:font-bold lg:font-bold" style="--tw-text-opacity: 1;color: ` +
			`rgb(239 68 68 / var(--tw-text-opacity));font-size: 1.25rem;line-height: ` +
			`1.75rem;">This is a test is it working</p></body>`,

		`<body><h1 class="md:font-bold" style="font-size: 1.5rem;line-height: ` +
			`2rem;">Hello </h1><p style="--tw-text-opacity: 1;color: ` +
			`rgb(239 68 68 / var(--tw-text-opacity));font-size: 1.25rem;line-height: ` +
			`1.75rem;">This is a test is it working</p></body>`}

	for i := range tpl {
		result := removeOldClassTags(tpl[i])

		if !strings.EqualFold(result, expected[i]) {
			t.Errorf("Incorrect Result, result: %s expected: %s", result, expected[i])
		}
	}
}

func TestTransform(t *testing.T) {
	classes := []string{"w-[600px]", "*:italic", "@xl:p-2"}

	expected := []string{`.w-\[600px\]`, `.\*\:italic`, `.\@xl\:p-2`}

	for i, class := range classes {
		result := transform(class)

		if !strings.EqualFold(result, expected[i]) {
			t.Errorf("Incorrect Result, result: %s expected: %s", result, expected[i])
		}
	}
}

func Test_TransformImgTags(t *testing.T) {
	tags := []string{`<img src="/public/images/동작절.jpg" alt="A temple surrounded by trees"></img>`,
		`<img src="/public/test.jpg" alt="A test image" style="padding-top: 0.5rem;"></img>`,
		`<img src="/test1.jpg" alt="1st test image"></img><img src="/test2.jpg" alt="2nd test image"></img>`,
		`<img src="/public/test.jpg" alt="A test image" style="padding-top: 0.5rem; display:block;"></img>`}

	expected := []string{`<img src="/public/images/동작절.jpg" alt="A temple surrounded by trees" style="display:block;"></img>`,
		`<img src="/public/test.jpg" alt="A test image" style="padding-top: 0.5rem; display:block;"></img>`,
		`<img src="/test1.jpg" alt="1st test image" style="display:block;"></img><img src="/test2.jpg" alt="2nd test image" style="display:block;"></img>`,
		`<img src="/public/test.jpg" alt="A test image" style="padding-top: 0.5rem; display:block;"></img>`}

	for i, v := range tags {
		result := TransformImgTags(v)

		if !strings.EqualFold(result, expected[i]) {
			t.Errorf("Incorrect Result, result: %s expected: %s", result, expected[i])
		}
	}
}
