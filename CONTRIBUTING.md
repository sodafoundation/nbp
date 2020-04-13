# NBP

## How to contribute

nbp is [Apache 2.0](LICENSE) licensed and accepts contributions via GitHub pull requests. This document outlines some of the conventions on commit message formatting, contact points for developers and other resources to make getting your contribution into nbp easier.

## Email and chat

- Email: [opensds-dev](https://groups.google.com/forum/?hl=en#!forum/opensds-dev)
- Slack: #[opensds](https://opensds.slack.com) 

Before you start, NOTICE that ```master``` branch is the relatively stable version
provided for customers and users. So all code modifications SHOULD be submitted to
```development``` branch.

## Getting started

- Fork the repository on GitHub.
- Read the README.md for project information and build instructions.

For those who just get in touch with this project recently, here is a proposed contributing [tutorial](https://github.com/leonwanghui/installation-note/blob/master/opensds_fork_contribute_tutorial.md).

## Contribution Workflow

### Code style

The coding style suggested by the Golang community is used in nbp. See the [doc](https://github.com/golang/go/wiki/CodeReviewComments) for more details.

Please follow this style to make nbp easy to review, maintain and develop.

### Report issues

A great way to contribute to the project is to send a detailed report when you encounter an issue. We always appreciate a well-written, thorough bug report, and will thank you for it!

When reporting issues, refer to this format:

- What version of env (nbp, os, golang etc) are you using?
- Is this a BUG REPORT or FEATURE REQUEST?
- What happened?
- What you expected to happen?
- How to reproduce it?(as minimally and precisely as possible)

### Propose blueprints

- Raise your idea as an [issues](https://github.com/sodafoundation/nbp/issues).
- After reaching consensus in the issue discussion, complete the development on the forked repo and submit a PR. 
  Here are the [PRs](https://github.com/sodafoundation/nbp/pulls?q=is%3Apr+is%3Aclosed) that are already closed.
- If a PR is submitted by one of the core members, it has to be merged by a different core member.
- After PR is sufficiently discussed, it will get merged, abondoned or rejected depending on the outcome of the discussion.
