# Toast Notifications Guide

## Overview
Toast notifications are now limited to **ERROR messages only** with helpful user guides for fixing issues.

## Design Philosophy
- ✅ **Show errors only** - Users need to know when something goes wrong
- ❌ **No success toasts** - Actions complete silently for a cleaner UX
- 📚 **Include fix guides** - Every error includes step-by-step resolution instructions
- ⏱️ **6-second duration** - Enough time to read the error and guides
- 🎨 **Consistent styling** - Red background (#FEF2F2) with red border (#FCA5A5)

---

## Error Scenarios

### 1. Document Upload Errors

#### Error: Missing Required Fields
**Trigger:** User tries to upload without selecting document type or file

**Toast Message:**
```
❌ Unable to upload document
• Select a document type from the dropdown
• Choose a file (PDF, JPG, or PNG)
• Maximum file size: 10MB
```

**Location:** `/components/UploadModal.tsx`

---

#### Error: File Size Too Large
**Trigger:** User tries to upload a file larger than 10MB

**Toast Message:**
```
❌ File size too large
• Maximum file size is 10MB
• Try compressing your PDF or reducing image quality
• For help, contact your consultant via Messaging
```

**Location:** `/components/UploadModal.tsx`

---

### 2. Appointment Scheduling Errors

#### Error: Missing Required Fields
**Trigger:** User tries to schedule without filling all fields (date, time, purpose)

**Toast Message:**
```
❌ Cannot schedule appointment
• Select a date from the calendar
• Choose an available time slot (Mon-Fri, 9 AM - 4 PM EST)
• Select a purpose for your call
• All fields are required
```

**Location:** `/components/ScheduleModal.tsx`

---

#### Error: Invalid Date (Past Date)
**Trigger:** User tries to schedule appointment in the past

**Toast Message:**
```
❌ Invalid appointment date
• Cannot schedule appointments in the past
• Please select today or a future date
• For urgent matters, contact your consultant via Messaging
```

**Location:** `/components/ScheduleModal.tsx`

---

### 3. Password Update Errors

#### Error: Missing Password Fields
**Trigger:** User tries to update password without filling all fields

**Toast Message:**
```
❌ Password update failed
• All password fields are required
• Enter your current password
• Create a new password (min. 12 characters)
• Confirm your new password
```

**Location:** `/components/pages/SettingsPage.tsx`

---

#### Error: Passwords Don't Match
**Trigger:** New password and confirmation don't match

**Toast Message:**
```
❌ Passwords do not match
• New password and confirmation must match
• Check for typos in both fields
• Password is case-sensitive
```

**Location:** `/components/pages/SettingsPage.tsx`

---

#### Error: Weak Password
**Trigger:** New password is less than 12 characters

**Toast Message:**
```
❌ Password too weak
• Password must be at least 12 characters
• Include uppercase and lowercase letters
• Add numbers and special characters
• Avoid common words or patterns
```

**Location:** `/components/pages/SettingsPage.tsx`

---

## Removed Success Toasts

The following success notifications have been **removed** for cleaner UX:

| Action | Previous Behavior | New Behavior |
|--------|------------------|--------------|
| Document Upload | ✅ "Document uploaded successfully" toast | Silent success, modal closes |
| Appointment Scheduled | ✅ "Appointment confirmed" toast | Silent success, modal closes |
| Profile Updated | ✅ "Profile updated successfully" toast | Silent success (auto-save) |
| Password Changed | ✅ "Password updated successfully" toast | Silent success, form clears |
| SMS Preferences Saved | ✅ "SMS preferences saved" toast | Silent success (auto-save) |
| Video Watched | ✅ "Video marked as completed" toast | Silent success, modal closes |
| Notification Bell Click | ℹ️ "No new notifications" toast | No action (removed) |

---

## Toast Configuration

### Error Toast Styling
```typescript
toast.error('Error Title', {
  description: '• Guide point 1\n• Guide point 2\n• Guide point 3',
  duration: 6000,
  style: {
    background: '#FEF2F2',
    border: '1px solid #FCA5A5',
    borderRadius: '10px',
  },
});
```

### Key Properties
- **background:** `#FEF2F2` (Light red)
- **border:** `1px solid #FCA5A5` (Red accent)
- **borderRadius:** `10px` (Matches design system)
- **duration:** `6000ms` (6 seconds)
- **description:** Multi-line bullet points with `\n` separators

---

## Implementation Details

### Files Modified

1. **`/components/UploadModal.tsx`**
   - Added file size validation (10MB limit)
   - Enhanced error messages with fix guides
   - Removed success toast

2. **`/components/ScheduleModal.tsx`**
   - Added past date validation
   - Enhanced error messages with fix guides
   - Removed success toast

3. **`/components/pages/SettingsPage.tsx`**
   - Added password validation (length, matching)
   - Added state management for password fields
   - Enhanced error messages with fix guides
   - Removed all success toasts

4. **`/components/VideoModal.tsx`**
   - Removed success toast
   - Removed unused toast import

5. **`/components/DashboardLayout.tsx`**
   - Removed info toast for notifications
   - Removed unused toast import

---

## User Experience Benefits

### ✅ Pros
- **Cleaner UI** - No unnecessary notifications interrupting workflow
- **Better error handling** - Users know exactly how to fix problems
- **Faster interactions** - No need to wait for success confirmations
- **Reduced notification fatigue** - Only important errors are shown
- **Guided troubleshooting** - Step-by-step instructions included

### 📋 Best Practices
- Errors are shown for **6 seconds** (enough time to read)
- Bullet points (`•`) make guides **scannable**
- Line breaks (`\n`) separate different tips
- **Actionable instructions** tell users exactly what to do
- **Context-aware** suggestions (e.g., "contact consultant via Messaging")

---

## Future Enhancements

### Potential Additions
1. **Network Errors** - "Unable to connect. Check your internet connection."
2. **Session Timeout** - "Session expired. Please log in again."
3. **File Type Errors** - "Unsupported file type. Use PDF, JPG, or PNG."
4. **API Errors** - "Server error. Try again or contact support."
5. **Validation Errors** - "Invalid email format. Use: name@example.com"

### Error Categories
- 🔴 **Critical** - System failures, data loss risks
- 🟠 **Warning** - User action needed, fixable issues
- 🔵 **Info** - Optional improvements, tips (future consideration)

---

## Testing Checklist

- [x] Document upload without file selected
- [x] Document upload with oversized file (>10MB)
- [x] Appointment scheduling with missing fields
- [x] Appointment scheduling with past date
- [x] Password update with missing fields
- [x] Password update with mismatched passwords
- [x] Password update with weak password (<12 chars)
- [x] All success actions complete silently
- [x] No info/success toasts appear

---

**Last Updated:** October 29, 2025  
**Toast Library:** Sonner v2.0.3  
**Design System:** AIO Fund Client Dashboard
