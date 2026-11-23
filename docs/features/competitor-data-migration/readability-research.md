# Research: explaining an automated migration agent

**Date:** 2026-11-01. **Purpose:** guide the research artifact’s information architecture and explanations. These findings inform a design hypothesis; they do not prove this artifact is understandable without representative-user evaluation.

## What readers must be able to answer

1. What would the migration agent do for a customer?
2. Which data could be transferred, and what happens to unsupported records?
3. How would it connect to each competitor?
4. How do AI planning and reliable execution work together?
5. Which limits require engineering, customer permission or a source/vendor decision?
6. What must the team build and verify before offering it?

Readability requires preserving these answers while organizing them around the reader’s questions. A word-count reduction alone cannot establish comprehension.

## Research and design implications

| Evidence | Finding relevant here | Application to the artifact |
|---|---|---|
| Mayer & Moreno (2003), *Nine Ways to Reduce Cognitive Load in Multimedia Learning*, Educational Psychologist 38(1), 43–52, DOI 10.1207/S15326985EP3801_6 | The paper synthesizes experiments on managing limited processing capacity, including segmentation, signaling and placing related explanatory material together. Its setting is multimedia learning, so applying it to a product research artifact is a reasoned design transfer. | Self-paced question-based views; one coherent explanation per view; labels next to each process stage; an example adjacent to the mechanism it explains. Avoid decorative movement and unrelated illustrations. [Paper, university-hosted copy](https://faculty.washington.edu/farkas/WDFR/MayerMoreno9WaysToReduceCognitiveLoad.pdf). |
| Amershi et al. (2019), *Guidelines for Human-AI Interaction*, CHI, DOI 10.1145/3290605.3300233 | The authors developed and evaluated 18 guidelines, including communicating capability and uncertainty, supporting correction and explaining behavior. | Separate observed capabilities, proposed agent behavior and unverified access. Show reasons for exceptions and what the customer would need to do. Do not present an illustrative run as working software. [Author-published paper](https://www.microsoft.com/en-us/research/wp-content/uploads/2019/01/Guidelines-for-Human-AI-Interaction-camera-ready.pdf). |
| Nielsen Norman Group, progressive disclosure | Core information must remain available up front; specialized detail can be secondary when the route is clearly labeled. | Keep the automatic-agent goal, source constraints and outcome visible. Put endpoint/code references behind “Technical detail,” rather than removing competitor or architectural content. [Industry guidance](https://www.nngroup.com/articles/progressive-disclosure/). |
| Nielsen Norman Group, heading-based scanning and information scent | Descriptive headings and link labels help readers identify where an answer belongs. | Use “What data can move?” and “How does the agent work?” instead of slogans or unexplained icons. Keep active navigation and the current question visible. [Scanning](https://www.nngroup.com/articles/layer-cake-pattern-scanning/), [information scent](https://www.nngroup.com/articles/information-scent/). |
| Nielsen Norman Group, comparison tables | Consistent attributes and concise, comparable content matter more than a decorative table. | Restore all ten competitors, compare connection route and available reads consistently, and label unknowns. Let a selected row reveal evidence and limits. [Comparison guidance](https://www.nngroup.com/articles/comparison-tables/). |
| W3C accessibility guidance | Normal text needs adequate contrast; readable blocks benefit from manageable line length, spacing and resizing/reflow. The visual-presentation guidance is Level AAA, not a blanket requirement for every interface. | Target at least 4.5:1 for ordinary text, allow browser/text-size changes, use left-aligned reading blocks around 60–70 characters, and verify reflow. A precise universal font size is not established by these sources. [Contrast](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html), [visual presentation](https://www.w3.org/WAI/WCAG21/Understanding/visual-presentation.html). |

## Proposed information architecture

- **The goal:** the automated customer experience and value, with a simple source → agent → ContentStudio explanation.
- **Data & limits:** select a data type to see automated work, required decisions and hard boundaries together.
- **Competitors:** all ten sources, with documented connection routes and a deeper Hootsuite assessment.
- **The agent:** follow discovery, mapping, transfer, verification and cutover; show the AI planner and execution workers separately.
- **Build & prove:** reusable platform components, new migration services, durable state and evidence required for release.
- **Evidence:** source documentation, HCI research, assumptions and downloads.

Each view should give an answer before its technical details. Core limitations remain visible alongside the capability they qualify. Avoid a single long summary, a sparse marketing page, and an undifferentiated research dump.

## Evaluation

Check whether a reader can locate and explain: social reconnection, an automatic connector for Hootsuite, the route for a source without a documented API, what AI is responsible for, what prevents duplicate publishing, and what “all data” can honestly mean. Validate navigation, keyboard operation, external links, text resizing, contrast and mobile layout. Browser tests can verify mechanics and rendering; they cannot certify user comprehension or business value.

## Industry guidance for the agent implementation

Anthropic’s engineering guidance distinguishes predefined workflows from model-directed agents and recommends adding complexity only when it improves the task. **Our inference:** use a durable transfer workflow with a constrained AI mapping planner. Copying bytes, verifying identifiers and enforcing permissions need explicit software guarantees. This does not require training a custom model initially. [Building effective agents](https://www.anthropic.com/research/building-effective-agents).
