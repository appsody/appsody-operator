# Appsody Application operator contributing guidelines
Welcome to the Appsody community!

You can contribute to the project in a variety of ways:

* Fixing or reporting bugs
* Improve documentation
* Contribute code
We welcome all contributions to the Appsody Application operator project and request you follow the guidelines below:


## Raising issues
A great way to contribute to the project is to raise a detailed issue when you encounter a problem.

Check that the list of project issues doesn't already include that problem or request before submitting an issue. If you find a matching issue, add a "+1" or comment indicating that you have the same issue, as this helps prioritize the most common problems and requests.

A good bug report is one that makes it easy for everyone to understand what you were trying to do and what went wrong. Provide as much context as possible so we can try to recreate the issue.

## Commit message guidelines
You should describe what changed and why.

To avoid duplication if you are making any breaking changes to the code base you MUST create a new GitHub issue to track the discussion. It is good practise to raise GitHub issues for fixes but if you prefer to just submit a pull request with your desired code changes then that is fine.

## Pull requests
If you're working on an existing issue, simply respond to the issue and express interest in working on it. This helps other people know that the issue is active, and hopefully prevents duplicated efforts.

To submit a proposed change:

* Fork the affected repository.
* Create a new branch for your changes.
* Develop the code/fix.
* Modify the documentation as necessary.
* Verify all CI status checks pass, and work to make them pass if failing.
The general rule is that all PRs should be 100% complete - meaning they should include documentation changes related to the change. A significant exception is work-in-progress PRs. These should be indicated by opening a draft pull request. To open a draft pull request, click the dropdown arrow that appears next to the “Create pull request” button and then select the "Create draft pull request" option.

## Contributor License Agreement
In order for us to merge any of your changes into Appsody Application operator, you need to sign the Contributor License Agreement. When you open up a Pull Request for the first time, a bot will comment with a link to the CLA. You can review or sign the CLA now here https://cla-assistant.io/appsody/appsody-operator

## Merge approval and release process
A maintainer may add "LGTM" (Looks Good To Me) or an equivalent comment to indicate that a PR is acceptable. Any change requires at least one LGTM. No pull requests can be merged until at least one maintainer signs off with an LGTM.

Once the PR has been merged, the release job will automatically run in the CI process of the specific repository. 

## License
This project is licensed under the Apache 2.0 license, and all contributed stacks must also be licensed under the Apache 2.0 license. Each contributed stack should include a LICENSE file containing the Apache 2.0 license. More information can be found in the [LICENSE](LICENSE) file or online at

http://www.apache.org/licenses/LICENSE-2.0