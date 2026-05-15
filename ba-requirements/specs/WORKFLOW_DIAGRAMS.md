# AIO Fund Manager - Workflow Diagrams

## Overview
This document contains detailed workflow diagrams for all major processes in the AIO Fund Manager platform.

---

## 1. Trust Activation Workflow - Nexxess Trust

```
┌─────────────────────────────────────────────────────────────┐
│                    START: Client Onboarding                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              System Receives Trust Application              │
│              Status: Not Started → In Progress              │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                 ┌────────────────┐
                 │  Task Type?    │
                 └────────────────┘
                   │    │    │    │
         ┌─────────┘    │    │    └─────────┐
         │              │    │              │
         ▼              ▼    ▼              ▼
    ┌────────┐   ┌────────┐   ┌────────┐   ┌────────┐
    │ Upload │   │ Review │   │Schedule│   │ Simple │
    └────────┘   └────────┘   └────────┘   └────────┘
         │              │          │            │
         │              │          │            │
         ▼              ▼          ▼            ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ Upload Panel │ │ Review Panel │ │Schedule Panel│ │ Simple Panel │
│              │ │              │ │              │ │              │
│ • Select Doc │ │ • Read Terms │ │ • Pick Date  │ │ • Show Info  │
│ • Add Notes  │ │ • E-Sign     │ │ • Select Time│ │ • Mark Done  │
│ • Submit     │ │ • Submit     │ │ • Confirm    │ │              │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
         │              │          │            │
         └──────────────┴──────────┴────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Validation  │
                  │   Required?  │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          │
          ┌──────────────┐    │
          │ Under Review │    │
          │   (Advisor)  │    │
          └──────────────┘    │
                    │          │
                    ▼          │
          ┌──────────────┐    │
          │  Approved?   │    │
          └──────────────┘    │
            │          │      │
           Yes        No      │
            │          │      │
            ▼          ▼      │
    ┌────────────┐ ┌────────────┐
    │ Completed  │ │ Rejected   │
    │            │ │ Resubmit   │
    └────────────┘ └────────────┘
            │          │      │
            └──────────┴──────┘
                    │
                    ▼
            ┌──────────────┐
            │ Update Task  │
            │   Status     │
            └──────────────┘
                    │
                    ▼
            ┌──────────────┐
            │  Send        │
            │ Notification │
            └──────────────┘
                    │
                    ▼
            ┌──────────────┐
            │ Update       │
            │ Timeline     │
            └──────────────┘
                    │
                    ▼
            ┌──────────────┐
            │ Calculate    │
            │ Progress %   │
            └──────────────┘
                    │
                    ▼
            ┌──────────────┐
            │ All Tasks    │
            │ Complete?    │
            └──────────────┘
                │        │
               Yes      No
                │        │
                ▼        │
        ┌────────────┐  │
        │  Trust     │  │
        │ Activated  │  │
        │            │  │
        │ Send Final │  │
        │   Docs     │  │
        └────────────┘  │
                │        │
                ▼        ▼
        ┌──────────────────┐
        │   Dashboard      │
        │   Updated        │
        └──────────────────┘
                │
                ▼
        ┌──────────────────┐
        │       END        │
        └──────────────────┘
```

### Trust Workflow - 22 Step Breakdown

**Phase 1: Initial Setup (Steps 1-5)**
1. Initial Consultation (Schedule)
2. Trust Type Selection (Simple)
3. Upload Trust Agreement (Upload)
4. Upload Identification Documents (Upload)
5. KYC Information Form (Review)

**Phase 2: Legal & Compliance (Steps 6-12)**
6. Beneficiary Information (Simple)
7. Trustee Designation (Review)
8. Asset Inventory Upload (Upload)
9. Financial Disclosure (Upload)
10. Tax Information (Simple)
11. Legal Review Scheduling (Schedule)
12. Sign Trust Agreement (Review)

**Phase 3: Funding & Activation (Steps 13-18)**
13. Funding Instructions (Simple)
14. Asset Transfer Documents (Upload)
15. Bank Account Setup (Schedule)
16. Review Final Documents (Review)
17. Compliance Verification (Simple)
18. Final Signature (Review)

**Phase 4: Completion (Steps 19-22)**
19. Trust Activation Confirmation (Simple)
20. Certificate of Trust (Upload)
21. Distribution Instructions (Simple)
22. Welcome Package & Next Steps (Simple)

---

## 2. Fund - Private Equity Workflow

```
┌─────────────────────────────────────────────────────────────┐
│              START: PE Fund Subscription                    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                Initial Consultation Scheduled               │
│                Panel: FundPEConsultationPanel               │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Investor Qualification Assessment              │
│              Panel: FundPEInvestorQualificationPanel        │
│              • Accreditation Status Check                   │
│              • Income/Net Worth Verification                │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Accredited? │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          ▼
          ┌──────────────┐  ┌────────────────┐
          │  Continue    │  │  Reject or     │
          │  Process     │  │  Request Docs  │
          └──────────────┘  └────────────────┘
                    │              │
                    ▼              │
┌─────────────────────────────────────────────────────────────┐
│                Upload Accreditation Documents               │
│                Panel: FundPEAccreditationDocumentsPanel     │
│                • CPA Letter                                 │
│                • Tax Returns                                │
│                • Broker Statements                          │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                Complete Investor Profile                    │
│                Panel: FundPEInvestorProfilePanel            │
│                • Personal Information                       │
│                • Investment Objectives                      │
│                • Risk Tolerance                             │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                Select Fund Structure                        │
│                Panel: FundPEFundStructureSelectionPanel     │
│                • Committed Capital Amount                   │
│                • Investment Period                          │
│                • Distribution Preferences                   │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Prepare Subscription Agreement                 │
│              Panel: FundPESubscriptionPreparationPanel      │
│              • Review Terms                                 │
│              • Confirm Commitment                           │
│              • Acknowledge Risks                            │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                Execute Subscription                         │
│                Panel: FundPESubscriptionExecutionPanel      │
│                • E-Signature Required                       │
│                • Wire Instructions Provided                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                Setup Capital Call Schedule                  │
│                Panel: FundPECapitalCallSetupPanel           │
│                • Review Schedule                            │
│                • Setup Auto-Notifications                   │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                Process Initial Capital Call                 │
│                Panel: FundPEInitialCapitalCallPanel         │
│                • Amount Due                                 │
│                • Due Date                                   │
│                • Payment Instructions                       │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Payment     │
                  │  Received?   │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          ▼
          ┌──────────────┐  ┌────────────────┐
          │  LP Activated│  │  Send Reminder │
          │              │  │  Wait for Pay  │
          └──────────────┘  └────────────────┘
                    │              │
                    └──────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Update Dashboard & Send Welcome Kit            │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                       END                                   │
│                 LP Successfully Onboarded                   │
└─────────────────────────────────────────────────────────────┘
```

### PE Fund Workflow - 22 Step Breakdown

**Phase 1: Qualification (Steps 1-6)**
1. Initial Consultation
2. Investor Qualification Questionnaire
3. Accreditation Verification
4. Upload Accreditation Documents
5. Background Check Consent
6. Investor Profile Completion

**Phase 2: Fund Selection (Steps 7-12)**
7. Fund Structure Selection
8. Review Fund Terms
9. Investment Amount Determination
10. Risk Assessment
11. Suitability Analysis
12. Fund Manager Meeting

**Phase 3: Documentation (Steps 13-18)**
13. Subscription Agreement Preparation
14. Operating Agreement Review
15. PPM Acknowledgement
16. Subscription Execution (E-Sign)
17. W-9/W-8 Tax Forms
18. Anti-Money Laundering Verification

**Phase 4: Funding (Steps 19-22)**
19. Capital Call Schedule Setup
20. Initial Capital Call
21. Wire Transfer Confirmation
22. LP Welcome Package

---

## 3. Fund - Real Estate Workflow

```
┌─────────────────────────────────────────────────────────────┐
│            START: Real Estate Fund Subscription             │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Initial Consultation Scheduled                 │
│              Panel: FundREInitialConsultationPanel          │
│              • Investment Goals Discussion                  │
│              • Property Type Preferences                    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Investor Qualification Assessment              │
│              Panel: FundREInvestorQualificationPanel        │
│              • Accreditation Check                          │
│              • Real Estate Experience                       │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Select Property Type Focus                     │
│              Panel: FundREPropertyTypeSelectionPanel        │
│              • Residential                                  │
│              • Commercial                                   │
│              • Industrial                                   │
│              • Mixed-Use                                    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Property    │
                  │  Available?  │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          ▼
          ┌──────────────┐  ┌────────────────┐
          │  Show        │  │  Add to        │
          │  Properties  │  │  Waitlist      │
          └──────────────┘  └────────────────┘
                    │              │
                    └──────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Prepare Subscription Documents                 │
│              Panel: FundRESubscriptionPreparationPanel      │
│              • Investment Amount                            │
│              • Hold Period                                  │
│              • Distribution Preferences                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Execute Subscription Agreement                 │
│              Panel: FundRESubscriptionExecutionPanel        │
│              • E-Sign Documents                             │
│              • Acknowledge Risk Factors                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Setup Capital Call Schedule                    │
│              Panel: FundRECapitalCallSchedulePanel          │
│              • Initial Draw Amount                          │
│              • Future Call Schedule                         │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Process Initial Capital Call                   │
│              Panel: FundREInitialCapitalCallPanel           │
│              • Payment Instructions                         │
│              • Escrow Account Details                       │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Capital     │
                  │  Received?   │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          ▼
          ┌──────────────┐  ┌────────────────┐
          │  Investment  │  │  Follow-up     │
          │  Active      │  │  Required      │
          └──────────────┘  └────────────────┘
                    │              │
                    └──────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Property Portfolio Access Granted              │
│              • Quarterly Reports                            │
│              • Property Updates                             │
│              • Distribution Schedule                        │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                       END                                   │
│              RE Fund Investor Activated                     │
└─────────────────────────────────────────────────────────────┘
```

### RE Fund Workflow - 22 Step Breakdown

**Phase 1: Discovery (Steps 1-6)**
1. Initial Consultation
2. Investment Goals Assessment
3. Investor Qualification
4. Property Type Selection
5. Market Analysis Review
6. Portfolio Strategy Discussion

**Phase 2: Due Diligence (Steps 7-12)**
7. Property Portfolio Review
8. Financial Projections Analysis
9. Risk Factor Disclosure
10. Property Inspection Schedule
11. Third-Party Reports Review
12. Q&A with Fund Manager

**Phase 3: Subscription (Steps 13-18)**
13. Subscription Agreement Preparation
14. Operating Agreement Review
15. Property PPM Acknowledgement
16. Subscription Execution
17. Tax Documentation
18. Title & Insurance Review

**Phase 4: Funding & Activation (Steps 19-22)**
19. Capital Call Schedule
20. Initial Capital Call
21. Escrow Funding
22. Investor Portal Activation

---

## 4. Document Upload & Approval Workflow

```
┌─────────────────────────────────────────────────────────────┐
│              START: Client Uploads Document                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Open Upload Task Panel                         │
│              Component: UploadTaskPanel                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Select Document Type                           │
│              • Trust Agreement                              │
│              • Identification                               │
│              • Financial Statement                          │
│              • Tax Document                                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Choose File from Device                        │
│              Supported: PDF, JPG, PNG, DOCX                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Validation  │
                  │  Passed?     │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          ▼
          ┌──────────────┐  ┌────────────────┐
          │  Upload to   │  │  Show Error    │
          │  Server      │  │  Message       │
          └──────────────┘  └────────────────┘
                    │              │
                    ▼              │
          ┌──────────────┐        │
          │  Show Upload │        │
          │  Progress    │        │
          └──────────────┘        │
                    │              │
                    ▼              │
          ┌──────────────┐        │
          │  Upload      │        │
          │  Complete    │        │
          └──────────────┘        │
                    │              │
                    └──────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Document Status: Uploaded                      │
│              Create Document Record                         │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Notify Advisor for Review                      │
│              Send Email + In-App Notification               │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Advisor Reviews Document                       │
│              (Operations Portal)                            │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Approved?   │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          ▼
          ┌──────────────┐  ┌────────────────┐
          │  Status:     │  │  Status:       │
          │  Approved    │  │  Rejected      │
          └──────────────┘  └────────────────┘
                    │          │
                    │          ▼
                    │    ┌────────────────┐
                    │    │  Add Rejection │
                    │    │  Comments      │
                    │    └────────────────┘
                    │          │
                    ▼          ▼
┌─────────────────────────────────────────────────────────────┐
│              Notify Client of Status                        │
│              Update Task Status                             │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  If Rejected │
                  │  Resubmit?   │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    │          ▼
                    │    ┌────────────────┐
                    │    │  Escalate to   │
                    │    │  Compliance    │
                    │    └────────────────┘
                    │          │
                    └──────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Update Timeline & Progress                     │
│              Mark Task as Complete                          │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                       END                                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. Appointment Scheduling Workflow

```
┌─────────────────────────────────────────────────────────────┐
│              START: Schedule Appointment                    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Open Scheduling Panel                          │
│              Component: SchedulingTaskPanel                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Select Appointment Type                        │
│              • Initial Consultation                         │
│              • Legal Review                                 │
│              • Fund Manager Meeting                         │
│              • Property Inspection                          │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Check Advisor Availability                     │
│              Display Available Time Slots                   │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Client Selects Date & Time                     │
│              • Date Picker                                  │
│              • Time Slot Selector                           │
│              • Duration Display                             │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Slot        │
                  │  Available?  │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          ▼
          ┌──────────────┐  ┌────────────────┐
          │  Continue    │  │  Show Alt      │
          │              │  │  Times         │
          └──────────────┘  └────────────────┘
                    │              │
                    └──────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Add Meeting Details (Optional)                 │
│              • Meeting Notes                                │
│              • Special Requests                             │
│              • Virtual/In-Person                            │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Confirm Appointment                            │
│              Review All Details                             │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Create Calendar Event                          │
│              Status: Scheduled                              │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Send Confirmations                             │
│              • Email to Client                              │
│              • Email to Advisor                             │
│              • Calendar Invite (.ics)                       │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Add to Dashboard Calendar                      │
│              Create Reminder Notifications                  │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Wait for Appointment Date                      │
│              Send Reminders (24h, 1h before)                │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Meeting     │
                  │  Completed?  │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          ▼
          ┌──────────────┐  ┌────────────────┐
          │  Mark        │  │  Cancelled or  │
          │  Complete    │  │  Rescheduled   │
          └──────────────┘  └────────────────┘
                    │              │
                    ▼              │
          ┌──────────────┐        │
          │  Update Task │        │
          │  Status      │        │
          └──────────────┘        │
                    │              │
                    └──────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Update Timeline & Progress                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                       END                                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 6. Client Portal User Journey

```
┌─────────────────────────────────────────────────────────────┐
│                    START: Visit Website                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  User Type?  │
                  └──────────────┘
              │           │           │
         New User   Returning   Guest
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Register    │ │  Login Page  │ │  Learning    │
    │  (External)  │ │              │ │  Center      │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              │           ▼           │
              │    ┌──────────────┐  │
              │    │  Authenticate│  │
              │    │              │  │
              │    └──────────────┘  │
              │           │           │
              └───────────┴───────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Dashboard Home                           │
│                    • Portfolio Overview                     │
│                    • Action Items (2 Urgent)                │
│                    • Recent Activity                        │
└─────────────────────────────────────────────────────────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  My          │ │  Documents   │ │  Messages    │
    │  Portfolios  │ │              │ │              │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Select      │ │  Upload/View │ │  Chat with   │
    │  Portfolio   │ │  Files       │ │  Advisor     │
    └──────────────┘ └──────────────┘ └──────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Portfolio Detail View                    │
│                    • Overview Tab                           │
│                    • Documents Tab (15 docs)                │
│                    • Timeline Tab (20 activities)           │
│                    • Compliance Checklist                   │
└─────────────────────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Compliance Checklist                     │
│                    22 Steps • Color-Coded Sections          │
└─────────────────────────────────────────────────────────────┘
              │
              ▼
    ┌──────────────────┐
    │  Select Task     │
    └──────────────────┘
              │
              ▼
    ┌──────────────────┐
    │  Open Task Panel │
    └──────────────────┘
              │
    ┌─────────┴─────────┐
    │                   │
    ▼                   ▼
┌────────┐        ┌────────┐
│Complete│        │ Save & │
│ Task   │        │ Close  │
└────────┘        └────────┘
    │                   │
    └─────────┬─────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Update Progress                          │
│                    • Task Status Updated                    │
│                    • Progress Bar Updated                   │
│                    • Timeline Entry Added                   │
│                    • Notification Sent                      │
└─────────────────────────────────────────────────────────────┘
              │
              ▼
    ┌──────────────────┐
    │  All Tasks Done? │
    └──────────────────┘
        │           │
       Yes         No
        │           │
        ▼           ▼
┌──────────────┐ ┌────────────────┐
│  Portfolio   │ │  Continue      │
│  Activated   │ │  Working       │
│              │ │                │
│  Show        │ │  Return to     │
│  Success     │ │  Dashboard     │
└──────────────┘ └────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────┐
│                       END                                   │
│                Successful Onboarding                        │
└─────────────────────────────────────────────────────────────┘
```

---

## 7. Branding Configuration Workflow (Operations Portal)

```
┌─────────────────────────────────────────────────────────────┐
│              START: Operations Portal Login                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Navigate to Client Portal Settings             │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Tab 1: Branding & Login Page                   │
└─────────────────────────────────────────────────────────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Upload      │ │  Customize   │ │  Edit Login  │
    │  Logo        │ │  Colors      │ │  Info Cards  │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Image File  │ │  Primary:    │ │  Card 1:     │
    │  Upload      │ │  #2563EB     │ │  Shield      │
    │              │ │              │ │  "Secure..."  │
    │              │ │  Secondary:  │ │              │
    │              │ │  #7C3AED     │ │  Card 2:     │
    │              │ │              │ │  Clock       │
    │              │ │  Accent:     │ │  "24/7..."    │
    │              │ │  #F59E0B     │ │              │
    │              │ │              │ │  Card 3:     │
    │              │ │              │ │  BookOpen    │
    │              │ │              │ │  "Easy..."    │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              └───────────┴───────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Tab 2: Learning Center                         │
└─────────────────────────────────────────────────────────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Manage      │ │  AI Assistant│ │  Social      │
    │  Videos      │ │  Toggle      │ │  Links       │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Add Video   │ │  Enable/     │ │  YouTube     │
    │  • Title     │ │  Disable     │ │  LinkedIn    │
    │  • Desc      │ │              │ │  Twitter     │
    │  • Category  │ │              │ │  Facebook    │
    │  • Thumbnail │ │              │ │              │
    │  • Duration  │ │              │ │  + Add URL   │
    │  • Status    │ │              │ │  + Order     │
    │  • Order     │ │              │ │              │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              └───────────┴───────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Save All Changes                               │
│              Validation & Confirmation                      │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Sync to Client Portal                          │
│              • Update BrandingContext                       │
│              • Apply CSS Variables                          │
│              • Refresh Client Portal                        │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Client Portal Updated                          │
│              • New Logo Displayed                           │
│              • Colors Applied                               │
│              • Login Cards Updated                          │
│              • Learning Videos Updated                      │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                       END                                   │
│              White-Label Configuration Complete             │
└─────────────────────────────────────────────────────────────┘
```

---

*Last Updated: January 2026*
*Version: 1.2*
*Platform: AIO Fund Manager - Nexxess Business Advisors*
