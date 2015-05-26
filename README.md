# Pivotal Roadmap Generator

This application connects to your pivotal tracker API and creates a nice visual roadmap.

### Assumptions

* __You use Epics to track features.__ - This app uses epics as the feature list to show on the roadmap
* __You have a release story for each epic__ - This app looks for the last release story for a given feature/epic and uses the due date of that release to calculate the scheduled release of that feature. If there is more than one release, it uses the last release for the epic
* __Start date is calculated by the start date of the sprint of the first story for a given epic__ - This one was hard to describe. Just read it 10 times and you'll get it ;-)


### Download

Coming soon... 

### Usage

First, set the following environment variables:

```
$ export TOKEN='your Pivotal Tracker API token'
$ export PROJECT_ID=99

```

Next, run the app

```Go
$ pivotal-roadmap
```

Pivotal-roadmap will create an HTML file with your roadmap!

