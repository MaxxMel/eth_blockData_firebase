func main() {

	ETHclientURL := "<mainnet.infura.io/ key >"

	firebaseURL := "<firebase proj link/.json>"
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Приложение запущено. Данные отправляются каждые 5 секунд...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка приложения.")
			return
		case <-ticker.C:
			blockInfo := GetLatestBlockInfo(ETHclientURL)

			if blockInfo.Error != nil {
				log.Printf("Ошибка: %v\n", blockInfo.Error)
				continue
			}

			if blockInfo.BlockHash == lastBlockHash {
				log.Printf("Блок с хешем %s уже был отправлен. Пропускаем отправку.\n", blockInfo.BlockHash)
				continue
			}

			err := uploadData(blockInfo, firebaseURL)
			if err != nil {
				log.Printf("Ошибка загрузки данных: %v\n", err)
			} else {
				log.Printf("Данные успешно отправлены: %+v\n", blockInfo)
			}
		}
	}
}
