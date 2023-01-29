# hcj

HCJ is an HTML/CSS and JS renderer.

## Status

As of January 2023, HCJ has semi-functional HTML and CSS, and no JS capabilities.

| Goal               | Status                               |
| ------------------ | ------------------------------------ |
| HTML Parsing       | Complete via `golang.org/x/net/html` |
| CSS Parsing        | ~95% complete                        |
| JS Parsing         | 0% complete                          |
| HTML/CSS Rendering | ~10% complete                        |
| JS Rendering       | 0% complete                          |
| Tests Passing      | 52                                   |
| Test Count         | 257                                  |

## Philosophy

This project is test driven, it will be considered complete when there are no test files run through its
test suites that fail to render a correct image or video output, and when these suites fully cover the feature
set of the target languages.

The input data for test suites can be found in `testdata`.

## Contributing

- Choose an existing test file from a `todo` directory or find your own. The former is more likely to be accepted as a non-duplicate test. Move this to a `in` testdata directory.
- Change the library such that the test suites produce a valid render of the selected test file.
- Ensure these changes do not change, or do not significantly change, or only correct existing mistakes in existing output files.
- Open a PR with the code and testdata changes.

## Notes

- HTML and CSS rendering are combined as HTML cannot be rendered without some concept of styles; style-less HTML is still driven by baked-in default styles, which are roughly agreed upon by existing web browsers.
- This project targets CSS3 and HTML5, but earlier versions of these languages should be targeted for intermediate functionality goals.
- Some functionality may be explicitly avoided to produce a library with less potential exploits, particularly around code execution, network access, and local file system access.
- To date this project is not interested in looking at existing open source implementations e.g. Chromium and copying their approach; this would save time at the cost of producing two identical implementations, with this project being likely inferior.
