# Code of Conduct

## Our Pledge

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone, regardless of age, body size, visible or invisible disability, ethnicity, sex characteristics, gender identity and expression, level of experience, education, socio-economic status, nationality, personal appearance, race, religion, or sexual identity and orientation.

We pledge to act and interact in ways that contribute to an open, welcoming, diverse, inclusive, and healthy community focused on advancing cybersecurity education and responsible security research.

## Our Standards

### Examples of behavior that contributes to a positive environment:

* **Demonstrating empathy and kindness** toward other people
* **Being respectful** of differing opinions, viewpoints, and experiences
* **Giving and gracefully accepting constructive feedback**
* **Accepting responsibility** and apologizing to those affected by our mistakes
* **Focusing on what is best** for the overall community and cybersecurity education
* **Using responsible disclosure** for security vulnerabilities
* **Following ethical guidelines** for security research

### Examples of unacceptable behavior:

* The use of sexualized language or imagery, and sexual attention or advances of any kind
* Trolling, insulting or derogatory comments, and personal or political attacks
* Public or private harassment
* Publishing others' private information without explicit permission
* **Using this framework for unauthorized or malicious purposes**
* **Sharing bypass techniques without corresponding defensive documentation**
* **Encouraging or assisting in illegal activities**
* Other conduct which could reasonably be considered inappropriate in a professional setting

## Enforcement Responsibilities

Project maintainers are responsible for clarifying and enforcing our standards of acceptable behavior and will take appropriate and fair corrective action in response to any behavior that they deem inappropriate, threatening, offensive, or harmful.

Project maintainers have the right and responsibility to remove, edit, or reject comments, commits, code, wiki edits, issues, and other contributions that are not aligned to this Code of Conduct, and will communicate reasons for moderation decisions when appropriate.

## Scope

This Code of Conduct applies within all community spaces, and also applies when an individual is officially representing the community in public spaces. Examples of representing our community include using an official e-mail address, posting via an official social media account, or acting as an appointed representative at an online or offline event.

## Enforcement

### Reporting

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported to the project maintainers via:

- GitHub Issues (for non-sensitive matters)
- Private email (for sensitive concerns - see SECURITY_POLICY.md)

All complaints will be reviewed and investigated promptly and fairly.

### Enforcement Guidelines

Project maintainers will follow these Community Impact Guidelines in determining the consequences for any action they deem in violation of this Code of Conduct:

#### 1. Correction

**Community Impact**: Use of inappropriate language or other behavior deemed unprofessional or unwelcome.

**Consequence**: A private, written warning from project maintainers, providing clarity around the nature of the violation and an explanation of why the behavior was inappropriate. A public apology may be requested.

#### 2. Warning

**Community Impact**: A violation through a single incident or series of actions.

**Consequence**: A warning with consequences for continued behavior. No interaction with the people involved, including unsolicited interaction with those enforcing the Code of Conduct, for a specified period of time. This includes avoiding interactions in community spaces as well as external channels like social media. Violating these terms may lead to a temporary or permanent ban.

#### 3. Temporary Ban

**Community Impact**: A serious violation of community standards, including sustained inappropriate behavior.

**Consequence**: A temporary ban from any sort of interaction or public communication with the community for a specified period of time. No public or private interaction with the people involved, including unsolicited interaction with those enforcing the Code of Conduct, is allowed during this period. Violating these terms may lead to a permanent ban.

#### 4. Permanent Ban

**Community Impact**: Demonstrating a pattern of violation of community standards, including sustained inappropriate behavior, harassment of an individual, or aggression toward or disparagement of classes of individuals.

**Consequence**: A permanent ban from any sort of public interaction within the community.

## Responsible Security Research

As this is a cybersecurity education project, additional standards apply:

### Required Conduct for Security Research

1. **Authorization Required**
   - Only test on systems you own or have explicit written permission to test
   - Document authorization scope clearly
   - Stay within authorized boundaries

2. **Ethical Disclosure**
   - Report vulnerabilities responsibly (see SECURITY_POLICY.md)
   - Allow reasonable time for fixes before public disclosure
   - Do not exploit vulnerabilities for personal gain

3. **Purple Team Approach**
   - All offensive capabilities must include defensive documentation
   - Consider blue team perspective in all contributions
   - Prioritize education over exploitation

4. **Legal Compliance**
   - Follow all applicable laws and regulations
   - Respect intellectual property rights
   - Do not contribute code that violates laws

5. **No Malicious Use**
   - Do not use this framework for unauthorized access
   - Do not distribute for malicious purposes
   - Do not create or share malware variants

### Examples of Responsible Contributions

✅ **Good**:
```
"I've added SMB beaconing with corresponding YARA detection rules
and Splunk queries. Documentation includes blue team mitigation
strategies. Tested only in my authorized lab environment."
```

✅ **Good**:
```
"Found a bug in the DNS resolver that could leak data. Opened
private issue per SECURITY_POLICY.md. Proposing fix with tests."
```

❌ **Bad**:
```
"Here's a 0-day exploit for bypassing EDR. No detection methods
included because we want this to be stealthy."
```

❌ **Bad**:
```
"Modified implant to exfiltrate credentials without encryption.
Tested on my employer's network without permission."
```

## Attribution

This Code of Conduct is adapted from the [Contributor Covenant](https://www.contributor-covenant.org/), version 2.1, with additional security research guidelines specific to cybersecurity education projects.

For answers to common questions about this code of conduct, see the FAQ at https://www.contributor-covenant.org/faq.

## Contact

- GitHub Issues: https://github.com/Raoof128/CRTIAE/issues
- Security: See SECURITY_POLICY.md
- General questions: GitHub Discussions

---

**By participating in this project, you agree to abide by this Code of Conduct and use this framework responsibly and ethically.**
