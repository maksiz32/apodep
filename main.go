package main

import (
	"bufio"
	"bytes"
	"fmt"
	"myiopkg"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func allApo(drvs []string) []string {
	var files []string
	for _, drive := range drvs {
		fmt.Printf("Поиск в папке: %s ", drive)
		file, err := myiopkg.IsApoWalkDir(drive)
		if err != nil {
			fmt.Printf("\nВНИМАНИЕ! '%s' - такой директории не существует\n", drive)
		} else if file != nil {
			fmt.Println("- найдено установленное ПО АРО3.")
		} else {
			fmt.Println("- АРО3 не обнаружено.")
		}
		files = append(files, file...)
	}
	return files
}
func runMeElevated() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		fmt.Println(err)
	}
}
func amAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		fmt.Println("Программа запущена не под администратором\nПерезапускаюсь")
		return false
	}
	fmt.Println("Программа запущена под администратором")
	return true
}
func argsWinPath(s string) (a2 []string) {
	a := strings.Split(s, "\\")
	if a[len(a)-1] == "" {
		a1 := make([]string, len(a)-1)
		copy(a1, a)
		a2 = []string{strings.Join(a1, "\\")}
	} else {
		a2 = []string{strings.Join(a, "\\")}
	}
	return
}

func main() {
	if !amAdmin() {
		runMeElevated()
	}
	var drivers2 []string
	drivers := myiopkg.GetDrivies()
	if len(os.Args) > 1 {
		fmt.Printf("\nЗапускаю поиск АПО3 в указанном Вами расположении: %s\n", strings.ToUpper(os.Args[1]))
		fmt.Println()
		time.Sleep(time.Second * 2)
		drivers2 = argsWinPath(strings.ToLower(os.Args[1]))
	} else {
		usrProf := os.Getenv("USERPROFILE")
		sysDrive := os.Getenv("SystemDrive")
		drivers2 = []string{sysDrive + "\\RGS", usrProf + "\\Desktop", usrProf + "\\Documents", usrProf + "\\Downloads"}
		fmt.Println("Запускаю поиск АПО3 в стандартных папках:")
		fmt.Println()
	}
	isApo := allApo(drivers2)
	fmt.Println()
	if isApo == nil {
		fmt.Println("АПО3 в указанном размещении не найдено\nЧто делать дальше?")
		time.Sleep(time.Second * 1)
		fmt.Println("Попробовать искать АПО3 на всех доступных дисках? ВНИМАНИЕ: это может занять много времени (более 1 часа).")
		//Получить ДА или НЕТ (да, нет, д, н, yes, no, y, n - в любом регистре)
		input := myiopkg.YesNo()
		if bytes.EqualFold([]byte(input), []byte("y")) || bytes.EqualFold([]byte(input), []byte("д")) || bytes.EqualFold([]byte(input), []byte("yes")) || bytes.EqualFold([]byte(input), []byte("да")) {
			time.Sleep(time.Second * 3)
			fmt.Println("\nПоиск АПО3 по всем доступным дискам запущен...")
			isApo = allApo(drivers)
		} else {
			fmt.Println("Программа завершается...")
			time.Sleep(time.Second * 3)
			os.Exit(0)
		}
	}
	if len(isApo) == 1 {
		time.Sleep(time.Second * 1)
		fmt.Println("Найдена одна установка АРО3")
		fmt.Println()
		time.Sleep(time.Second * 1)
		fmt.Println("Выключаю DEP для АПО3")
		fmt.Println()
		time.Sleep(time.Second * 3)

		//Добавить DEP для isApo[0]+"APO3.exe")
		s := isApo[0] + "\\APO3.exe"
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers`, registry.WRITE)
		if err != nil {
			fmt.Printf("Программа завершилась с ошибкой и будет выключена:\n%#v\n", err)
			fmt.Println()
			time.Sleep(time.Second * 1)
			fmt.Println("Попробуйте запустить программу с правами Администратора")
			time.Sleep(time.Second * 8)
			os.Exit(0)
		}
		defer k.Close()
		err = k.SetStringValue(s, "DisableNXShowUI")
		if err != nil {
			fmt.Printf("Программа завершилась с ошибкой и будет выключена:\n%#v\n", err)
			fmt.Println()
			time.Sleep(time.Second * 1)
			fmt.Println("Попробуйте запустить программу с правами Администратора")
			time.Sleep(time.Second * 8)
			os.Exit(0)
		}
		fmt.Println("DEP для АПО3 успешно выключен.")
		fmt.Println()
		time.Sleep(time.Second * 1)
		fmt.Println("Для применения изменений необходима перезагрузка компьютера.")
		fmt.Println()
		time.Sleep(time.Second * 1)
		fmt.Println("Компьютер будет перезагружен автоматически через 1 минуту")
		time.Sleep((time.Second * 5))
		cmd2 := exec.Command("shutdown", "/r")
		err = cmd2.Run()
		if err != nil {
			fmt.Printf("Программа завершилась с ошибкой и будет выключена:\n%#v\n", err)
			fmt.Println()
			time.Sleep(time.Second * 1)
			fmt.Println("Перезагрузите компьютер самостоятельно")
			time.Sleep(time.Second * 8)
			os.Exit(0)
		}
		os.Exit(0)
	} else if len(isApo) > 1 {
		fmt.Println("Найдено несколько установок APO3")
		fmt.Println()
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers`, registry.WRITE)
		if err != nil {
			fmt.Printf("Программа завершилась с ошибкой и будет выключена:\n%#v\n", err)
			fmt.Println()
			time.Sleep(time.Second * 1)
			fmt.Println("Попробуйте запустить программу с правами Администратора")
			time.Sleep(time.Second * 8)
			os.Exit(0)
		}
		defer k.Close()
		//Прогнать добавление DEP для всех установок АПО3
		for i := 0; i < len(isApo); i++ {
			fmt.Println("Выключаю DEP для", isApo[i]+"\\APO3.exe")
			fmt.Println()
			time.Sleep(time.Second * 1)
			err = k.SetStringValue(isApo[i]+"\\APO3.exe", "DisableNXShowUI")
			if err != nil {
				fmt.Printf("Программа завершилась с ошибкой и будет выключена:\n%#v\n", err)
				fmt.Println()
				time.Sleep(time.Second * 1)
				fmt.Println("Попробуйте запустить программу с правами Администратора")
				time.Sleep(time.Second * 8)
				os.Exit(0)
			}
		}
		fmt.Println("DEP для всех АПО3 успешно выключен.")
		time.Sleep(time.Second * 4)
		fmt.Println()
		fmt.Println("Для применения изменений необходима перезагрузка компьютера.")
		fmt.Println()
		time.Sleep(time.Second * 1)
		fmt.Println("Компьютер будет перезагружен автоматически через 1 минуту")
		time.Sleep((time.Second * 5))
		cmd2 := exec.Command("shutdown", "/r")
		err = cmd2.Run()
		if err != nil {
			fmt.Printf("Программа завершилась с ошибкой и будет выключена:\n%#v\n", err)
			fmt.Println()
			time.Sleep(time.Second * 1)
			fmt.Println("Перезагрузите компьютер самостоятельно")
			time.Sleep(time.Second * 8)
			os.Exit(0)
		}
		os.Exit(0)
	} else {
		fmt.Println()
		fmt.Println("АПО3 на компьютере не найдено.")
		fmt.Println()
		fmt.Print("Для завершения программы нажмите клавишу Enter ")
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadString('\n')
		os.Exit(0)
	}
}
