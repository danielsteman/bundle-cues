#Job: {
    name: string
}

#Pipeline: {
    name: [...string]
}

#Channel: {
    id: string
}

#Schema: {
    include: [...string]
    resources: {
        jobs: {
            [string]: #Job
        }
        pipelines: [...string]
    }
    // "a"!: string
    targets: {
        prod: {
            resources: {
                jobs!: {
                    [string]: {
                        webhook_notifications!: {
                            on_failure: {
                                [...#Channel]
                            }
                        }
                        tasks: _
                    }
                }
            }
        }
        ...
    }
}

#Schema
