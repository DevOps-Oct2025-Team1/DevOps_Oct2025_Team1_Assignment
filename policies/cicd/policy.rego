package cicd.secrets

deny[msg] {
  step := input.jobs[_].steps[_]
  contains(lower(step.run), "api_key")
  not contains(step.run, "secrets.")
  msg := "Hardcoded API key detected. Use GitHub Secrets."
}
