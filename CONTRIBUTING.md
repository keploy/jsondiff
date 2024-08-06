# Contributing to Keploy

Thank you for your interest in Keploy and for taking the time to contribute to this project. 🙌 Keploy is a project by developers for developers and there are a lot of ways you can contribute.

If you don't know where to start contributing, ask us on our [Slack channel](https://join.slack.com/t/keploy/shared_invite/zt-2dno1yetd-Ec3el~tTwHYIHgGI0jPe7A).

## Code of conduct

Contributors are expected to adhere to the [Code of Conduct](CODE_OF_CONDUCT.md).

## Prerequisites for the contributors

Contributors should have knowledge of git, go, and markdown for most projects since the project work heavily depends on them.
We encourage Contributors to set up JsonDiff for local development and play around with the code and tests to get more comfortable with the project. 

Sections

- <a name="contributing"> General Contribution Flow</a>
  - <a name="#commit-signing">Developer Certificate of Origin</a>
- <a name="contributing-keploy">JsonDiff Contribution Flow</a>

# <a name="contributing">General Contribution Flow</a>

## <a name="commit-signing">Signing-off on Commits (Developer Certificate of Origin)</a>

To contribute to this project, you must agree to the Developer Certificate of
Origin (DCO) for each commit you make. The DCO is a simple statement that you,
as a contributor, have the legal right to make the contribution.

See the [DCO](https://developercertificate.org) file for the full text of what you must agree to
and how it works [here](https://github.com/probot/dco#how-it-works).
To signify that you agree to the DCO for contributions, you simply add a line to each of your
git commit messages:

```
Signed-off-by: Jane Smith <jane.smith@example.com>
```

In most cases, you can add this signoff to your commit automatically with the
`-s` or `--signoff` flag to `git commit`. You must use your real name and a reachable email
address (sorry, no pseudonyms or anonymous contributions). An example of signing off on a commit:

```
$ commit -s -m “my commit message w/signoff”
```

To ensure all your commits are signed, you may choose to add this alias to your global `.gitconfig`:

_~/.gitconfig_

```
[alias]
  amend = commit -s --amend
  cm = commit -s -m
  commit = commit -s
```

# How to contribute ?

We encourage contributions from the community.

**Create a [GitHub issue](https://github.com/keploy/jsondiff/issues) for any changes beyond typos and small fixes.**

We review GitHub issues and PRs on a regular schedule.

To ensure that each change is relevant and properly peer reviewed, please adhere to best practices for open-source contributions.
This means that if you are outside the Keploy organization, you must fork the repository and create PRs from branches on your own fork.
The README in GitHub's [first-contributions repo](https://github.com/firstcontributions/first-contributions) provides an example.

## ## How to set up the JsonDiff locally?

1. Fork the repository

<br/>

2. Clone the repository with the following command. Replace the <GITHUB_USERNAME> with your username

```sh
git clone https://github.com/<GITHUB_USERNAME>/jsondiff.git
```

<br/>

3. Go into the directory containing the project and edit the changes.

## <a name="contributing-keploy">JsonDiff Contribution Flow</a>

JsonDiff is written in `Go` (Golang) and leverages Go Modules. Relevant coding style guidelines are the [Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments) and the _Formatting and style_ section of Peter Bourgon's [Go: Best
Practices for Production Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).

There are many ways in which you can contribute to JsonDiff.

# Contact

Feel free to join [slack](https://join.slack.com/t/keploy/shared_invite/zt-2dno1yetd-Ec3el~tTwHYIHgGI0jPe7A) to start a conversation with us.
