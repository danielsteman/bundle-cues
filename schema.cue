#NonEmptyID: string & != ""

#Job: {
  // Other fields are allowed (open struct by default)
  ...
  webhook_notifications!: {
    on_failure!: [{
      id!: #NonEmptyID
    }, ...]
  }
}

targets: {
  prod: {
    resources: {
      jobs: [string]: #Job
    }
  }
}

