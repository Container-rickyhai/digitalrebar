This is the node classifier daemon for Digital Rebar.

It listens for on_milestone events, and for each event it recieves it tests the event to see if it matches a set of user-provided rules, and for each rule it matches it takes appropriate action.

## Building the tool:

* Download and install glide: https://glide.sh/
* Pull in the proper dependencies: glide install
* Build the tool: go build

## Command line options:
* -debug

  Whether to run in debug mode.  If passed, classifier will log extra debug logging.
* -rules [path-to-rulefile]

  The rules that events should be matched against.  The contents of this file are described later.
* -endpoint [url]

  The API endpoint for Digital Rebar.
* -username

  The username to log in to Digital Rebar with.
* -password

  The password for the Digital Rebar user.
* -listen

  The address:port that the classifier should listen for events on.
* -testRules

  If set, the classifier will just test the rulefile for validity and exit.

## The rules file:

The rules file should contain a list of rules in YML format.  For a
simple example, refer to the test.yml file.  The rule file will be
reloaded if the classfier is sent a SIGHUP.

### An example rules file:

    ---
    - Name: 'test logging rule'
      WantsAttribs: []
      Matchers:
        - Enabled: true
        - Script: "/bin/true"
        - JSON:
            Selector: ':root .Attribs .number_of_drives:val(5)'
            SaveAs: 'driveCount'
      MatchActions:
        - Log: true
        - Script: "/usr/bin/env"
  
### Rule definition:

A Rule is composed of the following fields:

* Name

  The name of the rule.  If present, it must be unique across all the rules.
* Description

  A brief description of what the rule does.  For human use only.
* WantsAttribs

  A list of DigitalRebar attribs or node attributes that
  the rule Matchers will use to determine whether the Rule matches the
  Event.
* Matchers

  A list of Matchers that must match in order for the MatchActions to run.
* MatchActions

  A list of Actions that will be run if all the Matchers match.

#### Matchers:
Currently, the classifier knows about the following matchers:

* And

  Takes a list of Matchers, and matches iff all of them match.
* Or

  Takes a list of Matchers, and matches if any of them match.
* Not

  Takes a single Matcher, and matches if it does not match.
* Enabled

Takes true or false, and matches if it was passed true.
* JSON

  Takes a YAML object of the following format:
      
        ---
        Selector: ':root .Attribs .number_of_drives:val(5)'
        SaveAs: 'driveCount'
        PickResults:
          driveCount: 0
  
  The individual keys have the following meanings:
  
    * Selector
      
      A JSONSelect selector fragment http://jsonselect.org/#overview.  This selector will be matched against the JSON serialization of the RunContext that the rule is running in.
      The Selector must match at least one thing in the RunContext unless PickResults was also specified, in which case it must match all of the variables that PickResults was asked to fill in.  
    * SaveAs
    
      The name of the variable to save whatever the Selector picked. If the selector only found one thing, that object will be saved, otherwise an array comprised of all the found objects will be saved.
    * PickResults
      An object whose keys indicate variables to save and whose values indicate the index in the Selector results of the value that should be saved.  If PickResults and SaveAs refer to the same variable, PickResults wins.
      
          
* Script
  Takes a shell script, and matches if the shell script exits with a zero status.
The shell script will be passed the following extra environment variables:

    * CLASSIFIER_ATTRIBS
      A JSON blob of the attributes listed in the rule WantsAttribs
    * CLASSIFIER_EVENT
      The event in JSON format
    * REBAR_ENDPOINT
      The API endpoint that DigitalRebar lives on.
    * REBAR_KEY
      The username:password for DigitalRebar

#### Actions:
Currently, the classifier knows about 2 actions:

* Log
  Emit a logging message.
* Script
Takes a shell script, which will be passed the following extra environment variables:

    * CLASSIFIER_ATTRIBS
      A JSON blob of the attributes listed in the rule WantsAttribs
    * CLASSIFIER_EVENT
      The event in JSON format
    * REBAR_ENDPOINT
      The API endpoint that DigitalRebar lives on.
    * REBAR_KEY
      The username:password for DigitalRebar
  

