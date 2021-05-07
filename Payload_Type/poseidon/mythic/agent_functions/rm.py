from mythic_payloadtype_container.MythicCommandBase import *
import json


class RmArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {}

    async def parse_arguments(self):
        if len(self.command_line) > 0:
            if self.command_line[0] != "{":
                data = {
                    "host": "",
                    "path": "",
                    "file": self.command_line
                }
                self.command_line = json.dumps(data)
        else:
            raise Exception("No command line arguments")


class RmCommand(CommandBase):
    cmd = "rm"
    needs_admin = False
    help_cmd = "rm [path]"
    description = "Delete a file."
    version = 1
    supported_ui_features = ["file_browser:remove"]
    author = "@xorrior"
    argument_class = RmArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        try:
            tmp = json.loads(task.args.command_line)
            if tmp["path"] != "":
                task.display_params = tmp["path"] + "/" + tmp["file"]
            else:
                task.display_params = tmp["file"]
        except:
            pass
        return task

    async def process_response(self, response: AgentResponse):
        pass
