# von Explaino auto-diagram

Reads a logical/ physical application from iServer and guides turning it into a C4 diagram

## Components

* Main - just runs the thing
* Fyne - provides the GUI
* Search - searches iServer to find the base component
* Explorer - finds related components to add to the diagram
* Diagram - converts the found items into a C4 representation
* Visualise - displays the diagram and allows for tweaking

### Notes

#### Solution Concept Diagram

> A Solution Concept diagram provides a high-level orientation of the solution that is envisaged in order to meet the objectives of the architecture engagement. In contrast to the more formal and detailed architecture diagrams developed in the following phases, the solution concept represents a "pencil sketch" of the expected solution at the outset of the engagement.
> 
> This diagram may embody key objectives, requirements, and constraints for the engagement and also highlight work areas to be investigated in more detail with formal architecture modeling.
> 
> Its purpose is to quickly on-board and align stakeholders for a particular change initiative, so that all participants understand what the architecture engagement is seeking to achieve and how it is expected that a particular solution approach will meet the needs of the enterprise.
>
> - https://pubs.opengroup.org/architecture/togaf91-doc/arch/chap35.html

## Todo

* [x] Alter the Display screen to provide a different form for PAC, PTC, and LAC things.
* New ideas
  * [ ] Build a report to mirror the HERM
  * [ ] Use a Force Directed Graph to look for missing bits, see congreunce