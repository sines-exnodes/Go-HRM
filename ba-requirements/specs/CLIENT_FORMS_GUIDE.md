# 📋 Client-View Forms Implementation Guide

## Overview
This guide documents the 4 client-facing compliance forms that have been implemented in the AIO Fund Client Dashboard. These forms are simplified versions designed for client data entry, with "Submit for Review" buttons instead of approval workflows.

---

## 🎯 Form Types

### 1. 👥 Beneficiary Screening (Task ID: 3)

**Purpose:** Collect beneficiary details for verification and PEP screening

**Form Structure:**
- **Single-column layout** for optimal readability
- **2 sections:** Beneficiary Information + Screening Declaration

#### Section 1: Beneficiary Information
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Full Name | Text Input | ✅ Yes | Legal name of beneficiary |
| Date of Birth | Date Picker | ✅ Yes | DOB for identity verification |
| Relationship to You | Text Input | ✅ Yes | Relationship to settlor (e.g., Spouse, Child) |
| Country of Residence | Text Input | ✅ Yes | Current country of residence |
| Upload ID Document | File Upload | ✅ Yes | PDF or Image (MAX. 10MB) |

#### Section 2: Screening Declaration
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Is this person a PEP? | Dropdown | ✅ Yes | Options: No, Yes |
| Additional Notes | Textarea | ❌ No | Optional notes if PEP status is Yes |

**Validation Error:**
```
⚠️ Please complete all required fields before submitting

• Fill in beneficiary information (name, DOB, relationship, country)
• Upload ID document
• Answer PEP screening question
• All fields marked with * are required
```

**Button Text:** "Submit for Review"

---

### 2. 💳 Payment Confirmation (Task ID: 7)

**Purpose:** Confirm trust setup payment submission

**Form Structure:**
- **Single-column layout**
- **2 sections:** Payment Information + Confirmation

#### Section 1: Payment Information
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Invoice Number | Text Input | ❌ No | Invoice reference if provided |
| Payment Method | Dropdown | ✅ Yes | Options: Wire Transfer, Credit Card, Check, Other |
| Amount Paid (USD) | Currency Input | ✅ Yes | Dollar amount with $ prefix |
| Date of Payment | Date Picker | ✅ Yes | Date payment was made |
| Upload Payment Proof | File Upload | ❌ No | Optional receipt/confirmation (MAX. 10MB) |

#### Section 2: Confirmation
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Confirmation Checkbox | Checkbox | ✅ Yes | "I confirm that the above information is true and accurate" |

**Validation Error:**
```
⚠️ Please confirm payment details before submitting

• Select payment method
• Enter amount paid
• Select payment date
• Check the confirmation checkbox
• All fields marked with * are required
```

**Button Text:** "Submit for Review"

---

### 3. ⚖️ Legal Document Upload (Task ID: 8)

**Purpose:** Upload signed trust documents for legal review

**Form Structure:**
- **Single-column layout**
- **2 sections:** Document Uploads + Declaration

#### Section 1: Document Uploads
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Upload Trust Deed (Signed) | File Upload | ✅ Yes | Signed trust deed in PDF format |
| Upload Power of Attorney | File Upload | ❌ No | POA document if applicable |
| Upload Additional Legal Agreements | File Upload | ❌ No | Any other relevant legal documents |

**File Upload Requirements:**
- **Format:** PDF only
- **Size Limit:** 10MB per file
- **Drag & Drop:** Supported

#### Section 2: Declaration
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Confirmation Checkbox | Checkbox | ✅ Yes | "I confirm these are true and accurate copies" |

**Validation Error:**
```
⚠️ Please upload all required documents before submitting

• Upload signed trust deed
• Upload additional documents if applicable
• Check the confirmation checkbox
• All required fields must be completed
```

**Button Text:** "Submit for Review"

---

### 4. 🪙 Source of Funds Declaration (Task ID: 9)

**Purpose:** Declare fund sources for AML compliance

**Form Structure:**
- **Single-column layout**
- **3 sections:** Source of Funds + Wealth Overview + Acknowledgment

#### Section 1: Source of Funds
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Primary Source of Funds | Dropdown | ✅ Yes | Options: Employment Income, Business Profits, Inheritance, Investment Returns, Other |
| Brief Description | Textarea | ✅ Yes | How funds were earned (3 rows) |
| Upload Supporting Document | File Upload | ✅ Yes | Bank statement, contract, etc. (PDF/Image, MAX. 10MB) |

#### Section 2: Wealth Overview
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Approximate Net Worth (USD) | Currency Input | ✅ Yes | Total net worth with $ prefix |
| Country of Main Asset Holdings | Text Input | ✅ Yes | Country where most assets are held |

#### Section 3: Acknowledgment
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Confirmation Checkbox | Checkbox | ✅ Yes | "I declare that the information provided is accurate and truthful" |

**Validation Error:**
```
⚠️ Please fill in all required details before submitting

• Select primary source of funds
• Provide description of how funds were earned
• Upload supporting document
• Enter net worth and asset country
• Check the acknowledgment checkbox
```

**Button Text:** "Submit for Review"

---

## 🎨 Design System

### Form Styling (All Forms)

**Layout:**
- Single-column layout (no 2-column grids in client view)
- Section headers: `text-sm font-medium text-[#0B1930]`
- Spacing between sections: `space-y-4`
- Spacing within sections: `space-y-3`

**Input Fields:**
- Labels: `text-sm` (14px)
- Inputs: `h-9` (36px height), `text-sm`
- Margin top: `mt-1.5` (6px between label and input)
- Placeholder text: Gray (`text-gray-500`)

**File Upload Areas:**
- Height: `h-24` (96px)
- Border: `border-2 border-dashed`
- Background: `bg-gray-50` with `hover:bg-gray-100`
- Border color: `border-gray-300`
- Icon size: `w-6 h-6`
- Upload icon color: `text-gray-500`

**Checkboxes:**
- Label: `text-sm leading-snug cursor-pointer`
- Margin top for alignment: `mt-0.5`

**Currency Inputs:**
- Dollar sign prefix: Absolutely positioned at `left-3`
- Input padding: `pl-7` (to make room for $)

**Separators:**
- Between sections using `<Separator />` component

### Color System

| Element | Color | Hex |
|---------|-------|-----|
| Section Headers | Dark Navy | `#0B1930` |
| Required Field Labels | Default Text | Inherits from theme |
| Input Borders | Gray | From Shadcn |
| Upload Areas (Default) | Light Gray | `#F9FAFB` |
| Upload Areas (Hover) | Gray | `#F3F4F6` |
| Upload Border | Gray | `#D1D5DB` |
| Button Primary | Gold | `#C6A661` |
| Button Hover | Darker Gold | `#B39551` |
| Error Toast Background | Light Red | `#FEF2F2` |
| Error Toast Border | Red | `#FCA5A5` |

---

## 🔒 Validation Rules

### Common Validation Patterns

**All Required Fields:**
- Validated on submit button click
- Error toast displays for 6 seconds
- Submit button shows "Submitting..." during validation
- Form does not close if validation fails

**File Uploads:**
- Max size: 10MB per file
- Accepted formats vary by field (PDF only vs PDF/Images)
- No validation on actual file upload (client-side)
- Upload is visually indicated but not enforced in current implementation

**Checkboxes (Confirmations):**
- Must be checked before submission
- Boolean validation: `!formData.fieldName`

**Date Fields:**
- HTML5 date picker (no past date validation client-side)
- Format: YYYY-MM-DD

**Currency Fields:**
- Numeric input with step validation
- Dollar sign ($) prefix for visual clarity
- No decimal validation enforced (allows free input)

---

## 🔕 Toast Notification System

### Error Toasts Only (No Success Toasts)

Following the design philosophy:
- ✅ **Errors show** - With helpful guides for fixing issues
- ❌ **Success is silent** - Form closes smoothly on successful submission
- 📋 **Actionable guidance** - Bullet points tell users exactly what to fix

### Error Toast Configuration

```typescript
toast.error('⚠️ Error Title', {
  description: '• Guide point 1\n• Guide point 2\n• Guide point 3',
  duration: 6000,
  style: {
    background: '#FEF2F2',
    border: '1px solid #FCA5A5',
    borderRadius: '10px',
  },
});
```

**Properties:**
- **Duration:** 6 seconds (6000ms)
- **Background:** Light red (`#FEF2F2`)
- **Border:** Red accent (`1px solid #FCA5A5`)
- **Border Radius:** `10px`
- **Icon:** ⚠️ (warning emoji)

---

## 📂 File Structure

**Component:** `/components/ComplianceTaskDrawer.tsx`

**Form Functions:**
1. `BeneficiaryScreeningForm()` - Lines ~467-550
2. `PaymentConfirmationForm()` - Lines ~822-925
3. `LegalReviewForm()` - Lines ~927-1005
4. `SourceOfFundsForm()` - Lines ~1006-1121

**Validation Function:**
- `handleSubmit()` - Lines ~54-136
- Includes validation for tasks 3, 7, 8, 9

**Button Text Function:**
- `getButtonText()` - Lines ~169-177
- Returns "Submit for Review" for tasks 3, 7, 8, 9

---

## 🚀 Form Workflow

### User Journey

1. **Open Form:**
   - User clicks compliance task from Trust Details page
   - Right-side drawer opens (640px width)
   - Form loads with empty state

2. **Fill Form:**
   - User enters data in required fields (marked with *)
   - Optional fields can be skipped
   - File uploads use drag-and-drop or click
   - Data persists in component state as user types

3. **Submit Form:**
   - User clicks "Submit for Review" button
   - Button shows "Submitting..." state
   - Validation runs on client-side

4. **Validation Results:**
   
   **If validation fails:**
   - Error toast appears for 6 seconds
   - Bullet-point guide shows what to fix
   - Form stays open
   - Button re-enables
   
   **If validation passes:**
   - Silent success (no toast)
   - Form drawer closes smoothly
   - Task marked as complete
   - Trust progress updates
   - Activity log appends new entry

---

## 🧪 Testing Checklist

### Beneficiary Screening Form
- [ ] All required fields enforce validation
- [ ] PEP dropdown has correct options (No, Yes)
- [ ] File upload area displays correctly
- [ ] Error toast shows when fields missing
- [ ] Submit button says "Submit for Review"
- [ ] Silent success on valid submission

### Payment Confirmation Form
- [ ] Currency input shows $ prefix
- [ ] Payment method dropdown has 4 options
- [ ] Date picker works correctly
- [ ] Confirmation checkbox enforced
- [ ] Error toast shows when confirmation unchecked
- [ ] Optional fields can be skipped

### Legal Review Form
- [ ] Three separate file upload areas render
- [ ] PDF-only acceptance enforced (accept=".pdf")
- [ ] Declaration checkbox enforced
- [ ] Error toast shows when checkbox unchecked
- [ ] Submit button says "Submit for Review"

### Source of Funds Form
- [ ] Primary source dropdown has 5 options
- [ ] Textarea has 3 rows (not resizable)
- [ ] Net worth currency input has $ prefix
- [ ] Acknowledgment checkbox enforced
- [ ] All 3 sections display with separators
- [ ] Error toast comprehensive for all fields

---

## 📊 Data Flow

### Form Data Structure

**Beneficiary Screening (Task 3):**
```typescript
{
  beneficiaryName: string,
  beneficiaryDob: string,
  relationship: string,
  countryResidence: string,
  pepStatus: 'no' | 'yes',
  additionalNotes?: string
}
```

**Payment Confirmation (Task 7):**
```typescript
{
  invoiceNumber?: string,
  paymentMethod: 'wire-transfer' | 'credit-card' | 'check' | 'other',
  amountPaid: number,
  paymentDate: string,
  paymentConfirmation: boolean
}
```

**Legal Review (Task 8):**
```typescript
{
  legalConfirmation: boolean
  // File uploads handled separately
}
```

**Source of Funds (Task 9):**
```typescript
{
  primarySource: 'employment' | 'business' | 'inheritance' | 'investment' | 'other',
  briefDescription: string,
  netWorth: number,
  assetCountry: string,
  sofAcknowledgment: boolean
}
```

### onComplete Callback

When form is successfully submitted:
```typescript
onComplete(task.id, formData);
```

This triggers:
1. Task status update to "completed"
2. Trust progress recalculation
3. Activity log entry creation
4. Form data persistence
5. Drawer close animation

---

## 🎯 Key Differences from Admin Forms

| Feature | Admin Forms | Client Forms |
|---------|-------------|--------------|
| **Layout** | 2-column grids | Single-column only |
| **Approval Fields** | Checkboxes for approval | No approval fields |
| **Review Dropdowns** | Status dropdowns | Simple confirmations |
| **Internal Notes** | Admin-only notes fields | Client-facing only |
| **Button Text** | "Mark as Complete" | "Submit for Review" |
| **Validation** | Admin-controlled | Required field validation |
| **Field Count** | 8-12 fields | 5-8 fields (simplified) |

---

## 🔮 Future Enhancements

### Potential Additions

1. **File Upload Feedback:**
   - Show uploaded file name
   - Display file size
   - Allow file removal before submit

2. **Real-time Validation:**
   - Validate fields on blur
   - Show inline error messages
   - Highlight missing required fields

3. **Auto-save Draft:**
   - Save form data to localStorage
   - Restore on drawer reopen
   - Clear draft on successful submit

4. **Multi-file Upload:**
   - Support multiple file selection
   - Show list of uploaded files
   - Individual file remove buttons

5. **Progress Indicators:**
   - Show completion percentage
   - Visual progress bar
   - Required vs optional field counter

6. **Field Tooltips:**
   - Help icons with guidance
   - Example values
   - Format requirements

---

## 📞 Support & Troubleshooting

### Common Issues

**Q: Form won't submit even with all fields filled**
- **A:** Check that all confirmation checkboxes are checked. The validation requires boolean `true` values.

**Q: File upload area not responding**
- **A:** Ensure accept attribute matches file type. PDFs should use `accept=".pdf"`, images should use `accept=".pdf,image/*"`

**Q: Currency input shows incorrect formatting**
- **A:** Currency inputs use HTML5 number type with step validation. The $ is purely visual (CSS positioned).

**Q: Date picker shows different format**
- **A:** Date inputs use browser-native picker. Format may vary by browser/locale but value is always YYYY-MM-DD.

**Q: Error toast not showing**
- **A:** Check that Sonner is imported from `sonner@2.0.3` and toast styling matches the error format specification.

---

**Last Updated:** October 30, 2025  
**Component Version:** Client Forms v1.0  
**Status:** ✅ Production Ready  
**Design System:** AIO Fund Client Dashboard
