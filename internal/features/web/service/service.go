package web_service

type WebService struct {
	WebRepository WebRepository
}

type WebRepository interface {
	GetFile(filePath string) ([]byte, error)
}

func NewWebService(
	webRepository WebRepository,
) *WebService {
	return &WebService{
		WebRepository: webRepository,
	}
}