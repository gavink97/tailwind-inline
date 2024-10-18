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
	classes := []string{"text-2xl", "lg:p-8", "max-w-md", "p-8"}
	expected := []field{{Type: "inline", Class: "text-2xl"}, {Type: "media",
		Class: "lg:p-8"}, {Type: "inline", Class: "max-w-md"}, {Type: "inline",
		Class: "p-8"}}

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
	field := field{Type: "media", Class: "sm:text-black"}

	expected := `@media (min-width: 640px) {
    .sm\:text-black {
    --tw-text-opacity: 1;
    color: rgb(0 0 0 / var(--tw-text-opacity));
    }` + "\n}"

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
	tpl := `<h1 class="lg:text-2xl p-8 sm:text-black">Hello</h1>`

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

	expected := "<head><style>\n" + `
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
    }` + "\n</style></head>" + `<h1 class="lg:text-2xl p-8 sm:text-black">Hello</h1>`

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
