# Testing strategy

- After each set of file changes that would make up working git commit, run the full test suite first before reporting the summarized file changes.
- There should be 2 types of tests: 
1. Go Unit Tests that are usually located alongside the program code
2. Shell Tests that should reside in <root_folder>/tests and concentrate more on integration tests

- A central feature of the app is the shell autocomplete functionality that should be tested as good and as much as possible
