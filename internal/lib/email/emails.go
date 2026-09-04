package email

func (c *SMTPClient) SendConfirmationEmail(to string, bookingID string, totalPrice float64) error {

	data := map[string]any{
		"BookingID": bookingID,
		// "PropertyName": propertyName,
		// "StartDate":    startDate.Format("January 2, 2006"),
		// "EndDate":      endDate.Format("January 2, 2006"),
		// "Address":      address,
		// "TotalMembers": totalMembers,
		"TotalPrice": totalPrice,
	}

	return c.SendEmail(to, "Booking Confirmation", "success-booking", data)
}

func (c *SMTPClient) SendOTP(to string, otp int) error {
	return c.SendEmail(to, "OTP Verification", "send-otp", map[string]any{"OTP_CODE": otp})
}
