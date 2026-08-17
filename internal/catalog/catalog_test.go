package catalog

import (
	"strings"
	"testing"
	"time"

	"volunteertraining/internal/domain"
	"volunteertraining/internal/store"
)

func TestImportCSVAndFilterRequired(t *testing.T) {
	database, err := store.Open(t.TempDir() + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	actor := domain.User{ID: "trainer", Name: "Trainer", Email: "trainer@example.test", Password: "pass", Role: domain.RoleTrainer, Active: true, CreatedAt: time.Unix(1, 0).UTC()}
	if err := database.SaveUser(actor); err != nil {
		t.Fatal(err)
	}
	service := NewService(database, func() time.Time { return time.Unix(2, 0).UTC() })
	data := "id,title,role,duration,url,required,tip,description\nvid-1,Welcome,greeter,100,https://test/one,true,Smile,Intro\nvid-2,Optional,greeter,50,https://test/two,false,Note,Optional\n"
	result, err := service.ImportCSV(actor, strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if result.VideosImported != 2 || !result.Successful() {
		t.Fatalf("result=%+v", result)
	}
	items, err := service.Required(domain.RoleGreeter)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "vid-1" {
		t.Fatalf("items=%+v", items)
	}
}
