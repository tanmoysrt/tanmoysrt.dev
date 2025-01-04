---
title: "Swiftwave PaaS: One Year in Development"
description: "Write your description here"
date: "2024-06-24"
is_redirect: false
redirect_url:
---

From last year (July~August), I am working on an open source PaaS to deploy and manage applications easily on any VPS. I have a motive to create a solution which you once setup on your cloud, you will get same kind of experience like other platforms like Heroku, Railway or Render.

## Little bit about the project - SwiftWave

SwiftWave is a self-hosted lightweight PaaS solution to deploy and manage your applications on any VPS.

Website - https://swiftwave.org/

Github - https://github.com/swiftwave-org/swiftwave

If you like the intiative give it a ⭐ star in GitHub.

## Why did I decide to start developing this open source PaaS?

**Incentive Motive** - I am absolutely exhausted from doing LeetCode. I don't usually enjoy CP or Leetcode, but need to do for placements. I began questioning if I really had a problem being consistent for long time (at-least 6 months). I therefore made an effort to push myself to create a project that I will stick with for a very long period.

**Passive Motive** - I like development and computer science core concepts so much and on every week build something for just fun or for hackathon. Anyway I don't want to spare money for deployment, so used Heroku, then moved to AWS EC2 (free tier) after getting familarized with linux. Many times my friends reach out to me for deploying in EC2.

While I was learning k8s and preparing for CKA, I amazed by it's internals & started thinking to build something lightweight with cool UI/UX. That's where it all get started.

**Target of Building Swiftwave** -

- **Being lightweight** : Swiftwave + Postgres + HAProxy + UDP Proxy = ~180MB of system. So leaving a lot of room for your deployed apps.
- **Scalable** : This isn't meant for huge-scale operations - Kubernetes handles that. Instead, it's aimed at meeting the needs of small startups, individual users, home labs, and internal company tools.
- **Full-Featured**: Swiftwave offers a wide range of capabilities. Visit [swiftwave.org](https://swiftwave.org/) to explore all the features available in the platform.

## Sprint vs Continuous Development

Before this project, I always go in Hackathon Mode for building project, giving all my free time. After launching the project, I often lost the motivation to keep maintaining it.

With Swiftwave, I started the same way but quickly realized this approach wouldn't work long-term. So, I switched it. For the first time, I tried a more structured approach, dedicating a consistent 1 to 1.5 hours daily to the project. This method proved effective. After a year of steady effort, we've successfully released a stable version of Swiftwave.

> A small effort every day builds more than occasional bursts of productivity.

## DSA is not useless

While I initially expressed disinterest in LeetCode, I don't find Data Structures and Algorithms (DSA) boring or unimportant. Many people think DSA is only useful for job interviews, but that's not the case.

Over the past year, I've applied DSA concepts in various parts of my projects. This experience has shown me that DSA knowledge is crucial in software development. However, it's worth noting that real-world problem-solving often takes more time than the quick solutions expected in interview.

For instance, in Swiftwave, we implemented a feature for uploading code directly to deploy apps. On the frontend, we create a tar archive of the project to send to the server. For NodeJS apps, we needed to exclude the 'node_modules' and other folder by following .gitignore rules.

Our first attempt at this was inefficient, taking about 30 seconds just to scan files, apply ignore rules, and create the archive.

![Worst Performance of tart](/assets/swiftwave-paas-one-year-in-development/worst-performance-of-tart.png)

In the end, we used a tree-based method to organize files and handle .gitignore rules. This new approach was much faster, finishing everything in about half a second."

![Superb Performance of tartplus](/assets/swiftwave-paas-one-year-in-development/superb-performance-of-tart.png)

**Source Code of tartplus** - https://github.com/swiftwave-org/tartplus

## Automation is a Must

For those starting an open source project with long-term maintenance in mind, automation is crucial. Initially, it's often a one person project or the work of a very small team.

Alongside maintaining a OSS project, maintainers have their full time job or academics. So, It's vital to automate all repetitive tasks, allowing maintainers to focus solely on coding and reviews.

In swiftwave, we have various things -

- \*.swiftwave.xyz DNS Resolver
- APT Repository - http://deb.repo.swiftwave.org/
- RPM Repository - http://rpm.repo.swiftwave.org/
- Custom build images for some one-click apps - https://github.com/swiftwave-org/app-store
- Deployment of documentation and websites

All these processes are automated.

For instance, I usually do a weekly release. I only need to draft and publish the release. Building APT & RPM packages for different architectures and uploading them to the repository server are all automated.

We primarily use Github Actions for nearly all our automation needs.

## Testcases boring but gives peace of mind

When I first started contributing to open source last year, I was assigned in some issues for writing test cases for software. I also had to write integration tests during my GSoC'23 project with CircuitVerse. Back then, I found writing tests tedious.

However, now that I'm developing my project, my perspective has changed. I've realized how crucial it is to write tests and increase code coverage. With good coverage, I don't need to manually check if features are affected by changes.

For software like Swiftwave, which users install on their own servers, a buggy release could seriously disrupt their systems. It's essential to ensure nothing breaks in stable releases. Having numerous test cases makes this process smoother and more efficient.

## It's not bad at all to reinvent the wheel

Sometimes, creating your own solution can be beneficial. Swiftwave needed a pubsub system and queueing mechanism. While RabbitMQ and Redis are common choices, they use up valuable server resources, leaving less for deployed apps.

My solution? I built an in-memory pubsub and persistent task queueing system with using goroutines. This approach uses minimal resources. As a result, Swiftwave typically runs using only about 45MB of RAM.

Swiftwave with all this things takes ~45MB of ram most of the time.

However, we've kept it flexible. Users can still configure Swiftwave to use Redis or RabbitMQ instead of the built-in system if they prefer.

## Changes in POV of Open Source

In the 'Indian YouTube Space', there's often too much focus on just contributing to open source, which might have changed what open source is really about. I used to think this way too.

Contributing to open source shouldn't feel forced. Here's a better approach:

- Start by using open source software
- Keep using it if you like it
- If you find an issue or want a new feature, try to help fix or add it
- Create your own open source projects that can help others
- If a project dependency isn't working well, try to fix its issues
- Engage in conversations and forums, helping others who are stuck

Project Maintainers, Creator become more happy when you use their products.

## Why Open Source code matters ?

Swiftwave is primarily built on two key softwares: Docker Swarm and HAProxy.

To develop various features, I had to understand how these tools work. This led me to explore several repositories:

- Docker Engine - https://github.com/moby/moby
- Docker CLI - https://github.com/docker/cli
- Libnetwork - https://github.com/moby/libnetwork
- HAProxy DataplaneAPI - https://github.com/haproxytech/dataplaneapi

One particular challenge was implementing direct SSH support for applications. I delved into the code for 'docker exec', which was quite complex. After grasping the concept, I was able to implement a similar feature in Swiftwave.

## The End

Developing Swiftwave has been an exciting and challenging journey that I want to continue long-term. It has been a great learning experience for me.

There's still lot of rooms for improvement in Swiftwave to make it an even more dependable deployment tool. We have many areas we'd like to enhance in the future.

---

Website - https://swiftwave.org/

Github - https://github.com/swiftwave-org/swiftwave

If you like the initiative give it a ⭐ star in GitHub.
