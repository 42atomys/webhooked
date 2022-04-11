---
name: Bug report
description: File a bug report
title: 'fix: '
labels: ['type/bug 🔥']
assignees: ''
body:
  # Describe
  - type: markdown
    id: describe
    attributes:
      label: Describe the bug
      description: Thanks for taking the time to describe the bug with more details as you can
      placeholder: "A bug happened when I receive data"
  # Reproduce
  - type: markdown
    id: reproduce
    attributes:
      label: To Reproduce
      description: Steps to reproduce the behavior
      placeholder: |
        1. My config is '...'
        2. Try to store following payload '....'
        3. See error
  # What happened ?
  - type: textarea
    id: what-happened
    attributes:
      label: What happened
      description: Also tell us, what did you expect to happen?
      placeholder: Tell us what you see!
      value: "A bug happened!"
    validations:
      required: true
  # Expected
  - type: textarea
    id: expected
    attributes:
      label: Expected behavior
      description: A clear and concise description of what you expected to happen, if applicable
      placeholder: Maybe can response with..
    validations:
      required: false
  # Version
  - type: dropdown
    id: version
    attributes:
      label: Version
      description: What version of our software are you running?
      options:
        - '1.0'
        - '0.5'
    validations:
      required: true
  # Logs
  - type: textarea
    id: logs
    attributes:
      label: Relevant log output
      description: Please copy and paste any relevant log output. This will be automatically formatted into code, so no need for backticks.
      render: shell
  - type: checkboxes
    id: terms
    attributes:
      label: Code of Conduct
      description: By submitting this issue, you agree to follow our [Code of Conduct](https://example.com)
      options:
        - label: I agree to follow this project's Code of Conduct
          required: true
  # Environment
  - type: dropdown
    id: environment
    attributes:
      label: Environment
      description: What is the environment used to run our software?
      options:
        - 'Docker'
        - 'Binary on Windows'
        - 'Binary on Darwin (osx)'
        - 'Binary on Linux (Ubuntu/Debin/Arch/...)'
    validations:
      required: true
  # Environment details
  - type: input
    id: environment-details
    attributes:
      label: Environment version
      description: If you use the binary version of software, please provide the version of your OS
      placeholder: OSX 10.11 (El Capitan)
    validations:
      required: false
  # additional context
  - type: markdown
    id: additional-context
    attributes:
      label: Additional context
      description: Add any other context about the problem here
