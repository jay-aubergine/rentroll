package rlib

import "time"

// DeleteDeposit deletes the Deposit associated with the supplied id
// For convenience, this routine calls DeleteDepositParts. The DepositParts are
// tightly bound to the Deposit. If a Deposit is deleted, the parts should be deleted as well.
func DeleteDeposit(id int64) {
	_, err := RRdb.Prepstmt.DeleteDeposit.Exec(id)
	if err != nil {
		Ulog("Error deleting Deposit for DID = %d, error: %v\n", id, err)
	}
}

// DeleteDepository deletes the Depository associated with the supplied id
func DeleteDepository(id int64) {
	_, err := RRdb.Prepstmt.DeleteDepository.Exec(id)
	if err != nil {
		Ulog("Error deleting Depository where DEPID = %d, error: %v\n", id, err)
	}
}

// DeleteDepositParts deletes ALL the DepositParts associated with the supplied id
func DeleteDepositParts(id int64) {
	_, err := RRdb.Prepstmt.DeleteDepositParts.Exec(id)
	if err != nil {
		Ulog("Error deleting DepositParts where DID = %d, error: %v\n", id, err)
	}
}

// DeleteInvoice deletes the Invoice associated with the supplied id
// For convenience, this routine calls DeleteInvoiceAssessments. The InvoiceAssessments are
// tightly bound to the Invoice. If a Invoice is deleted, the parts should be deleted as well.
// It also updates
func DeleteInvoice(id int64) error {
	_, err := RRdb.Prepstmt.DeleteInvoice.Exec(id)
	if err != nil {
		Ulog("Error deleting Invoice for InvoiceNo = %d, error: %v\n", id, err)
		return err
	}
	return DeleteInvoiceAssessments(id)
}

// DeleteInvoiceAssessments deletes ALL the InvoiceAssessments associated with the supplied InvoiceNo
func DeleteInvoiceAssessments(id int64) error {
	_, err := RRdb.Prepstmt.DeleteInvoiceAssessments.Exec(id)
	if err != nil {
		Ulog("Error deleting InvoiceAssessments where InvoiceNo = %d, error: %v\n", id, err)
	}
	return err
}

// DeleteJournalAllocations deletes the allocation records associated with the supplied jid
func DeleteJournalAllocations(jid int64) {
	_, err := RRdb.Prepstmt.DeleteJournalAllocations.Exec(jid)
	if err != nil {
		Ulog("Error deleting Journal allocations for JID = %d, error: %v\n", jid, err)
	}
}

// DeleteJournalEntry deletes the Journal record with the supplied jid
func DeleteJournalEntry(jid int64) {
	_, err := RRdb.Prepstmt.DeleteJournalEntry.Exec(jid)
	if err != nil {
		Ulog("Error deleting Journal entry for JID = %d, error: %v\n", jid, err)
	}
}

// DeleteJournalMarker deletes the JournalMarker record for the supplied jmid
func DeleteJournalMarker(jmid int64) {
	_, err := RRdb.Prepstmt.DeleteJournalMarker.Exec(jmid)
	if err != nil {
		Ulog("Error deleting Journal marker for JID = %d, error: %v\n", jmid, err)
	}
}

// DeleteLedgerEntry deletes the LedgerEntry record with the supplied lid
func DeleteLedgerEntry(lid int64) error {
	_, err := RRdb.Prepstmt.DeleteLedgerEntry.Exec(lid)
	if err != nil {
		Ulog("Error deleting LedgerEntry for LEID = %d, error: %v\n", lid, err)
	}
	return err
}

// DeleteLedger deletes the GLAccount record with the supplied lid
func DeleteLedger(lid int64) error {
	_, err := RRdb.Prepstmt.DeleteLedger.Exec(lid)
	if err != nil {
		Ulog("Error deleting GLAccount for LID = %d, error: %v\n", lid, err)
	}
	return err
}

// DeleteLedgerMarker deletes the LedgerMarker record with the supplied lmid
func DeleteLedgerMarker(lmid int64) error {
	_, err := RRdb.Prepstmt.DeleteLedgerMarker.Exec(lmid)
	if err != nil {
		Ulog("Error deleting LedgerMarker for LEID = %d, error: %v\n", lmid, err)
	}
	return err
}

// DeleteNote deletes the Note record with the supplied nid
func DeleteNote(nid int64) error {
	_, err := RRdb.Prepstmt.DeleteNote.Exec(nid)
	if err != nil {
		Ulog("Error deleting Note for NID = %d, error: %v\n", nid, err)
	}
	return err
}

// DeleteNoteType deletes the NoteType record with the supplied nid
func DeleteNoteType(nid int64) error {
	_, err := RRdb.Prepstmt.DeleteNoteType.Exec(nid)
	if err != nil {
		Ulog("Error deleting NoteType for NID = %d, error: %v\n", nid, err)
	}
	return err
}

// DeleteReceipt deletes the Receipt record with the supplied rcptid
func DeleteReceipt(rcptid int64) error {
	_, err := RRdb.Prepstmt.DeleteReceipt.Exec(rcptid)
	if err != nil {
		Ulog("Error deleting Receipt for RCPTID = %d, error: %v\n", rcptid, err)
	}
	return err
}

// DeleteReceiptAllocations deletes ReceiptAllocation records with the supplied rcptid
func DeleteReceiptAllocations(rcptid int64) error {
	_, err := RRdb.Prepstmt.DeleteReceiptAllocations.Exec(rcptid)
	if err != nil {
		Ulog("Error deleting ReceiptAllocation for RCPTID = %d, error: %v\n", rcptid, err)
	}
	return err
}

// DeleteCustomAttribute deletes CustomAttribute records with the supplied cid
func DeleteCustomAttribute(cid int64) error {
	_, err := RRdb.Prepstmt.DeleteCustomAttribute.Exec(cid)
	if err != nil {
		Ulog("Error deleting CustomAttribute for cid = %d, error: %v\n", cid, err)
	}
	return err
}

// DeleteCustomAttributeRef deletes CustomAttributeRef records with the supplied cid
func DeleteCustomAttributeRef(elemid, id, cid int64) error {
	_, err := RRdb.Prepstmt.DeleteCustomAttributeRef.Exec(elemid, id, cid)
	if err != nil {
		Ulog("Error deleting elemid=%d, id=%d, cid=%d, error: %v\n", elemid, id, cid, err)
	}
	return err
}

// DeleteRentableTypeRef deletes RentableTypeRef records with the supplied rid, dtstart and dtstop
func DeleteRentableTypeRef(rid int64, dtstart, dtstop *time.Time) error {
	_, err := RRdb.Prepstmt.DeleteRentableTypeRef.Exec(rid, dtstart, dtstop)
	if err != nil {
		Ulog("Error deleting RentableTypeRef with rid=%d, dtstart=%s, dtstop=%s, error: %v\n",
			rid, dtstart.Format(RRDATEINPFMT), dtstop.Format(RRDATEINPFMT), err)
	}
	return err
}

// DeleteRentableSpecialtyRef deletes RentableSpecialtyRef records with the supplied rid, dtstart and dtstop
func DeleteRentableSpecialtyRef(rid int64, dtstart, dtstop *time.Time) error {
	_, err := RRdb.Prepstmt.DeleteRentableSpecialtyRef.Exec(rid, dtstart, dtstop)
	if err != nil {
		Ulog("Error deleting RentableSpecialtyRef with rid=%d, dtstart=%s, dtstop=%s, error: %v\n",
			rid, dtstart.Format(RRDATEINPFMT), dtstop.Format(RRDATEINPFMT), err)
	}
	return err
}

// DeleteRentableStatus deletes RentableStatus records with the supplied rid, dtstart and dtstop
func DeleteRentableStatus(rid int64, dtstart, dtstop *time.Time) error {
	_, err := RRdb.Prepstmt.DeleteRentableStatus.Exec(rid, dtstart, dtstop)
	if err != nil {
		Ulog("Error deleting RentableStatus with rid=%d, dtstart=%s, dtstop=%s, error: %v\n",
			rid, dtstart.Format(RRDATEINPFMT), dtstop.Format(RRDATEINPFMT), err)
	}
	return err
}

// DeleteRentalAgreementPet deletes the pet with the specified petid from the database
func DeleteRentalAgreementPet(petid int64) error {
	_, err := RRdb.Prepstmt.DeleteRentalAgreementPet.Exec(petid)
	if err != nil {
		Ulog("Error deleting petid=%d error: %v\n", petid, err)
	}
	return err
}

// DeleteAllRentalAgreementPets deletes the pet with the specified petid from the database
func DeleteAllRentalAgreementPets(raid int64) error {
	_, err := RRdb.Prepstmt.DeleteAllRentalAgreementPets.Exec(raid)
	if err != nil {
		Ulog("Error deleting pets for rental agreement=%d error: %v\n", raid, err)
	}
	return err
}
