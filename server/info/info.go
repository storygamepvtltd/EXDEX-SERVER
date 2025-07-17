package info

import (
	"fmt"

	"github.com/fatih/color"
)

func ServerInfoInit() {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println(green(`
	
			███████╗██╗  ██╗██████╗ ███████╗██╗  ██╗    ██████╗ ██████╗ ███╗   ███╗
			██╔════╝╚██╗██╔╝██╔══██╗██╔════╝╚██╗██╔╝   ██╔════╝██╔═══██╗████╗ ████║
			█████╗   ╚███╔╝ ██║  ██║█████╗   ╚███╔╝    ██║     ██║   ██║██╔████╔██║
			██╔══╝   ██╔██╗ ██║  ██║██╔══╝   ██╔██╗    ██║     ██║   ██║██║╚██╔╝██║				
			███████╗██╔╝ ██╗██████╔╝███████╗██╔╝ ██╗██╗╚██████╗╚██████╔╝██║ ╚═╝ ██║
			╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═════╝ ╚═╝     ╚═╝
                                                    
	`))

	fmt.Println(cyan("\n🚀 Welcome to EXDEX! Fast & Powerful API server started... 💥\n"))
}
