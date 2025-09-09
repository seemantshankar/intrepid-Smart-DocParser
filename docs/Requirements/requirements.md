# Requirements Document

## Introduction

The Contract Analysis and Milestone Extraction system is a standalone
microservice built with strict coding best practices that automatically
analyzes uploaded contracts between buyers and sellers to extract contractual
obligations that trigger money transfers. The system leverages external LLMs
(like OpenAI GPT-4o) to identify, categorize, and sequence payment milestones
while providing risk analysis and vulnerability assessments. This microservice
exposes REST APIs for the main Smart Payment Infrastructure to consume,
enabling intelligent milestone-based Smart Cheque creation through
well-defined service contracts.

## Requirements

### Requirement 1

**User Story:** As a business user, I want to upload contract documents in
various formats with automatic validation, so that the system can analyze
contractual obligations without manual data entry or processing incorrect
documents.

#### Acceptance Criteria for Requirement 1

1. WHEN a user uploads a document THEN the system SHALL accept PDF, DOCX, TXT,
JPG, PNG, and TIFF file formats
2. WHEN a document is uploaded THEN the system SHALL validate file size limits
(max 10MB) and format compatibility
3. WHEN image files are uploaded THEN the system SHALL apply OCR processing
using cloud vision models like Qwen to extract text content
4. WHEN scanned PDFs are uploaded THEN the system SHALL detect if OCR is needed
and apply text extraction using cloud vision models like Qwen
5. WHEN text extraction is complete THEN the system SHALL use LLM analysis to
determine if the document is a valid contract
6. WHEN contract validation runs THEN the system SHALL identify key contract
elements (parties, obligations, terms, signatures/execution clauses)
7. IF the document is not a contract THEN the system SHALL notify the user with
specific reasons and suggest uploading a proper contract document
8. WHEN text extraction is complete THEN the system SHALL store the original
document and extracted text securely with confidence scores from the cloud
vision model
9. IF document upload fails THEN the system SHALL provide clear error messages
and retry options

### Requirement 2

**User Story:** As a contract analyst, I want the system to extract key contract
information and payment obligations, so that I have a complete overview of the
agreement and no financial commitments are missed.

#### Acceptance Criteria for Requirement 2

1. WHEN contract text is processed THEN the system SHALL use external LLM APIs
(GPT-4o, Claude, etc.) to analyze content
2. WHEN LLM analysis is performed THEN the system SHALL extract and display a
contract summary containing buyer name, seller name, nature of goods
(physical/digital/services), and total contract value
3. WHEN contract summary is generated THEN the system SHALL validate that all
required fields are populated and flag any missing critical information
4. WHEN LLM analysis is performed THEN the system SHALL identify all clauses
that trigger monetary transfers
5. WHEN payment obligations are found THEN the system SHALL extract obligation
descriptions, trigger conditions, and payment amounts
6. WHEN percentage-based payments are identified THEN the system SHALL calculate
absolute amounts based on total contract value from the summary
7. IF LLM analysis fails THEN the system SHALL retry with alternative models
and log failures for manual review

### Requirement 3

**User Story:** As a project manager, I want payment milestones organized in
chronological and functional order, so that I can understand the payment flow
throughout the contract lifecycle.

#### Acceptance Criteria for Requirement 3

1. WHEN payment obligations are extracted THEN the system SHALL sequence them
chronologically based on contract timeline
2. WHEN chronological ordering is complete THEN the system SHALL group related
obligations by functional categories
3. WHEN milestones are organized THEN the system SHALL assign percentage
allocations that sum to 100% of contract value
4. WHEN milestone sequencing is complete THEN the system SHALL validate logical
dependencies between milestones
5. IF milestone dependencies conflict THEN the system SHALL flag inconsistencies
for human review

### Requirement 4

**User Story:** As a risk manager, I want the system to identify missing
contractual elements and vulnerabilities based on industry best practices using
a comprehensive knowledge database, so that I can address gaps before contract
execution.

#### Acceptance Criteria for Requirement 4

1. WHEN contract analysis is complete THEN the system SHALL identify the nature
of buyer's and seller's businesses from contract content
2. WHEN business types are identified THEN the system SHALL determine the
relevant industry categories (manufacturing, services, technology, etc.)
3. WHEN industry context is established THEN the system SHALL first query its
internal industry knowledge database for best practices and standard
contractual terms
4. WHEN industry is not found in database THEN the system SHALL perform web
searches for current industry best practices and store findings in the industry
knowledge database
5. WHEN vulnerability assessment runs THEN the system SHALL flag potential risks
for both buyer and seller based on stored industry-specific standards
6. WHEN risk analysis is performed THEN the system SHALL suggest specific
contractual improvements aligned with database-stored industry best practices
7. WHEN recommendations are generated THEN the system SHALL categorize them by
severity (low, medium, high, critical) and include industry-specific reasoning
with database references
8. WHEN analysis is complete THEN the system SHALL provide actionable
recommendations with legal reasoning and industry benchmark references from the
knowledge database

### Requirement 5

**User Story:** As a system administrator, I want the system to maintain an
up-to-date industry knowledge database, so that contract analysis remains
current with evolving regulations and best practices.

#### Acceptance Criteria for Requirement 5

1. WHEN new industry findings are discovered through web searches THEN the
system SHALL store them in the industry knowledge database with proper
categorization and timestamps
2. WHEN storing industry knowledge THEN the system SHALL organize data by
industry type, jurisdiction, regulation type, and contractual best practices
3. WHEN the system runs periodic updates THEN it SHALL automatically refresh
industry knowledge database with latest regulations, laws, and contractual best
practices
4. WHEN periodic updates run THEN the system SHALL schedule them based on
industry volatility (high-regulation industries updated more frequently)
5. WHEN database updates occur THEN the system SHALL maintain version history
and track changes to industry standards over time
6. WHEN conflicting information is found during updates THEN the system SHALL
flag discrepancies for manual review and resolution
7. WHEN industry knowledge is accessed THEN the system SHALL log usage patterns
to optimize update frequencies and identify knowledge gaps

### Requirement 6

**User Story:** As a Smart Cheque creator, I want extracted milestones
automatically converted to Smart Cheque configurations, so that I can create
milestone-based payments without manual setup.

#### Acceptance Criteria for Requirement 6

1. WHEN milestone extraction is approved by both parties THEN the system SHALL
generate multiple Smart Cheque configurations based on the contractual
obligations
2. WHEN Smart Cheque configs are created THEN each SHALL include payer_id,
payee_id, amount, currency (USDT/USDC/e₹), milestones array, and contract_hash
3. WHEN milestone configs are generated THEN each milestone SHALL include id,
description, amount, verification_method (oracle/manual/hybrid), oracle_config,
sequence_order, dependencies, and trigger_conditions
4. WHEN Smart Cheques are created THEN they SHALL start in "created" status
(proposed state) and remain inactive until both parties approve
5. WHEN both parties approve THEN Smart Cheques SHALL transition through valid
states: created → locked → in_progress → completed (or disputed at any stage)
6. IF milestone mapping fails THEN the system SHALL provide manual configuration
options with pre-filled suggestions from the contract analysis

### Requirement 7

**User Story:** As a compliance officer, I want contract analysis results to
include regulatory and legal compliance checks, so that contracts meet
jurisdictional requirements.

#### Acceptance Criteria for Requirement 7

1. WHEN contracts are analyzed THEN the system SHALL check for required legal
clauses based on jurisdiction
2. WHEN compliance analysis runs THEN the system SHALL identify missing
regulatory requirements
3. WHEN legal gaps are found THEN the system SHALL suggest standard clause
additions
4. WHEN cross-border contracts are detected THEN the system SHALL flag
multi-jurisdictional compliance requirements
5. WHEN compliance check is complete THEN the system SHALL generate compliance
reports for legal review

### Requirement 8

**User Story:** As a contract party, I want the system to recommend optimal
dispute resolution pathways based on existing dispute handling infrastructure,
so that both parties have agreed-upon mechanisms for conflict resolution before
issues arise.

#### Acceptance Criteria for Requirement 8

1. WHEN contract analysis is complete THEN the system SHALL analyze the contract
type, value, and complexity to suggest appropriate dispute resolution
mechanisms using the existing ResolutionRoutingService
2. WHEN dispute pathways are generated THEN the system SHALL recommend resolution
methods from the existing system: mutual_agreement, mediation, arbitration,
court, administrative
3. WHEN dispute resolution recommendations are created THEN the system SHALL use
existing categorization rules to determine priority levels (low, normal, high,
urgent) and categories (payment, milestone, contract_breach, fraud, technical,
other)
4. WHEN mediation options are suggested THEN the system SHALL leverage existing
dispute management infrastructure including evidence handling, audit trails,
and notification systems
5. WHEN both parties approve dispute pathways THEN the system SHALL integrate
these mechanisms into Smart Cheque configurations with proper dispute status
transitions (initiated → under_review → escalated → resolved → closed)
6. WHEN Smart Cheques enter disputed status THEN they SHALL automatically freeze
funds and trigger the approved dispute resolution workflow
7. IF either party rejects the dispute pathways THEN the system SHALL allow
negotiation and modification until both parties reach agreement

### Requirement 9

**User Story:** As a business user, I want to visualize and edit milestone
workflows using a drag-and-drop interface, so that I can easily understand and
modify the payment sequence.

#### Acceptance Criteria for Requirement 9

1. WHEN AI analysis is complete THEN the system SHALL generate a visual
flowchart representation of milestones using Mermaid-style diagrams
2. WHEN users view the workflow THEN they SHALL see a drag-and-drop editor
similar to Zapier/Make/n8n interfaces
3. WHEN users modify the workflow THEN they SHALL be able to drag milestones to
reorder, edit conditions, and adjust payment amounts
4. WHEN modifications are made THEN the system SHALL validate workflow logic and
ensure total percentages equal 100%
5. WHEN edits are saved THEN the system SHALL track changes and maintain
detailed audit trails

### Requirement 10

**User Story:** As a contract party, I want collaborative approval of workflow
modifications, so that both buyer and seller must agree to changes before they
take effect.

#### Acceptance Criteria for Requirement 10

1. WHEN workflow modifications are made THEN the system SHALL notify the other
party via email with change details
2. WHEN notifications are sent THEN they SHALL include visual diff comparisons
showing before/after workflow states
3. WHEN the other party reviews changes THEN they SHALL be able to approve,
reject, or suggest counter-modifications
4. WHEN both parties approve THEN the system SHALL commit changes and update the
Smart Cheque configuration
5. IF either party rejects changes THEN the system SHALL revert to the previous
approved version and allow further negotiation

### Requirement 11

**User Story:** As a contract party, I want the system to break approved
contracts into multiple Smart Cheques automatically with integrated dispute
handling, so that payments are executed as milestones are completed throughout
the contract lifecycle.

#### Acceptance Criteria for Requirement 11

1. WHEN final contract workflow and dispute pathways are approved by both parties
THEN the system SHALL create individual Smart Cheques for each milestone
2. WHEN Smart Cheques are generated THEN the total amount SHALL be distributed
across milestones according to the approved percentage allocations
3. WHEN Smart Cheques are created THEN each SHALL include verification_method
mapping to existing systems (logistics APIs for delivery, manual approval for
quality checks, oracle integration for automated verification)
4. WHEN Smart Cheques are created THEN each SHALL include the approved dispute
resolution pathways and automatic integration with the existing dispute
management system
5. WHEN disputes arise on Smart Cheques THEN the system SHALL automatically
create dispute records with proper categorization, priority assignment, and
evidence handling capabilities
6. WHEN Smart Cheques enter "disputed" status THEN funds SHALL be frozen and the
dispute SHALL follow the pre-approved resolution pathway (mutual_agreement →
mediation → arbitration → court as configured)
7. WHEN disputes are resolved THEN Smart Cheques SHALL automatically transition
to appropriate final states based on resolution outcomes (completed for payment
release, or other states based on dispute resolution)
8. WHEN Smart Cheques are in "created" status THEN they SHALL remain inactive and
no funds SHALL be locked until payer activates them
9. WHEN payer activates Smart Cheques THEN they SHALL transition to "locked"
status with funds secured in XRPL escrow addresses

### Requirement 12

**User Story:** As a system administrator, I want comprehensive logging and
monitoring of contract analysis processes, so that I can ensure system
reliability and troubleshoot issues.

#### Acceptance Criteria for Requirement 12

1. WHEN contract analysis starts THEN the system SHALL log all processing steps
with timestamps
2. WHEN LLM API calls are made THEN the system SHALL log request/response data
and performance metrics
3. WHEN errors occur THEN the system SHALL capture detailed error information
and context
4. WHEN analysis is complete THEN the system SHALL store confidence scores and
processing metadata
5. WHEN monitoring alerts trigger THEN the system SHALL notify administrators of
system issues or degraded performance

### Requirement 13

**User Story:** As a developer, I want the system built with strict coding best
practices and clean architecture, so that the codebase is maintainable,
scalable, and follows industry standards.

#### Acceptance Criteria for Requirement 13

1. WHEN code is written THEN it SHALL follow clean architecture principles with
clear separation of concerns
2. WHEN implementing functionality THEN the system SHALL use SOLID principles
and dependency inversion patterns
3. WHEN creating components THEN they SHALL be modular with well-defined
interfaces and contracts
4. WHEN writing code THEN it SHALL maintain 95%+ test coverage with
comprehensive unit and integration tests
5. WHEN committing code THEN it SHALL pass automated linting, formatting, and
security scanning checks
6. WHEN building services THEN they SHALL implement proper error handling,
logging, and observability patterns
7. WHEN integrating external services THEN they SHALL use resilience patterns
like circuit breakers and retry logic

### Requirement 14

**User Story:** As a system architect, I want a robust HTTP server foundation
with production-ready middleware, so that the microservice can handle
enterprise-grade traffic securely and reliably.

#### Acceptance Criteria for Requirement 14

1. WHEN the HTTP server starts THEN it SHALL implement graceful shutdown and
proper signal handling
2. WHEN requests are received THEN the system SHALL apply comprehensive
middleware including CORS, rate limiting, request logging, and recovery
3. WHEN processing requests THEN the system SHALL validate input/output with
structured error responses and proper HTTP status codes
4. WHEN handling authentication THEN the system SHALL implement JWT-based
authentication and role-based authorization middleware
5. WHEN tracing requests THEN the system SHALL propagate correlation IDs and
implement distributed tracing
6. WHEN monitoring health THEN the system SHALL provide health check endpoints
with dependency validation
7. WHEN serving APIs THEN the system SHALL generate and serve OpenAPI
documentation with proper versioning
