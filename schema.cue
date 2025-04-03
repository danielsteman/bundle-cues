#NonEmptyID: string & != ""

#Job: {
    webhook_notifications: {
        on_failure: [{
            id: #NonEmptyID
            ...
        }, ...]
        ...
    }
    ...
}

targets: {
    prod: {
        resources: {
            jobs: [string]: #Job
        }
    }
}

