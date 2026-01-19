package kubernetes.images

allowed_images := {
  "alpine",
  "ubuntu",
  "debian",
}

deny[msg] {
  image := input.spec.template.spec.containers[_].image
  endswith(image, ":latest")
  msg := sprintf("Image '%s' uses :latest tag", [image])
}

deny[msg] {
  from := input.stages[_].from
  not allowed_images[from]
  msg := sprintf("Base image '%s' is not approved", [from])
}

