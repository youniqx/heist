package random

func (r *randomAPI) GenerateRandomString(length int) (string, error) {
	log := r.Core.Log().WithValues("method", "GenerateRandomString", "length", length)

	randomBase64String, err := r.fetchRandomBase64String(length)
	if err != nil {
		log.Info("failed to fetch random base64 encoded string", "error", err)
		return "", err
	}

	return randomBase64String[0:length], nil
}
