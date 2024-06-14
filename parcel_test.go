package main

import (
	"database/sql"
	"github.com/google/uuid"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	id, _ := uuid.NewRandom()

	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Number:    int(id.ID()),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// Prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	require.NoError(t, err)
	require.NotZero(t, id)

	// Get
	p, err := store.Get(id) // получите только что добавленную посылку, убедитесь в отсутствии ошибки
	require.NoError(t, err)
	// Проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	require.Equal(t, parcel, p)

	// Delete
	err = store.Delete(id) // удалите добавленную посылку, убедитесь в отсутствии ошибки
	require.NoError(t, err)
	// Проверьте, что посылку больше нельзя получить из БД
	_, err = store.Get(id)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	// Получите добавленную посылку и убедитесь, что адрес обновился
	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, p.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")

	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set status
	newStatus := "sent"
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента

func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	// check
	require.Len(t, storedParcels, len(parcels))

	for _, parcel := range storedParcels {
		// убедимся, что все посылки из storedParcels есть в parcelMap
		expectedParcel, ok := parcelMap[parcel.Number]
		require.True(t, ok)

		// убедимся, что значения полей полученных посылок заполнены верно
		assert.Equal(t, expectedParcel, parcel)
	}
}
