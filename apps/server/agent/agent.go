package agent

import (
	"log"

	"server/agent/models"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"
)

func CreateJAJAAgent() (*agent.Agent, error) {

	model, err := models.NewAnthropicModel("claude-sonnet-4-6")

    if err != nil {
        log.Fatalf("Failed to create model: %v", err)
    }

    tools, err := getTools()

    if err != nil {
        log.Fatalf("Failed to get tools: %v", err)
        return nil, err
    }

    // 2. Define the agent.
    jaja, err := llmagent.New(llmagent.Config{
        Name:        "jaja_agent",
        Model:       model,
        Description: "An academic assignment completion agent that reads assignment files and instructions, then produces complete, submission-ready Word documents.",
        Instruction: `You are JAJA (Just Automate Junk Assignments), an expert academic assistant specializing in completing school assignments to a high standard.
            You will be given assignment files (PDFs, images) and instructions describing what needs to be completed. Your job is to:
            1. Carefully read and analyze all provided assignment materials and instructions.
            2. Produce a thorough, well-structured response that fully satisfies the assignment requirements.
            3. Match the academic level, tone, and formatting conventions expected for the course.
            4. Use the docx tool to generate the final submission as a Word document.

            Always complete the full assignment — do not summarize or skip sections. If the assignment requires specific formatting (headings, citations, word count), follow it exactly.`,
        Tools: tools,
    })

    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
        return nil, err
    }

    return &jaja, nil

    // 3. Configure the launcher and run.
    // config := &launcher.Config{
    //     AgentLoader: agent.NewSingleLoader(a),
    // }

    // l := full.NewLauncher()
    // if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
    //     log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
    // }
}

func getTools() ([]tool.Tool, error) {
    // TODO: uncomment in next PR
    // return all tools here
    // docxTool, err := tools.DocxTool()
    // if err != nil {
    //     log.Fatalf("Failed to create docx tool: %v", err)
    //     return nil, err
    // }

    // return []tool.Tool{docxTool}, nil
    return []tool.Tool{}, nil
}