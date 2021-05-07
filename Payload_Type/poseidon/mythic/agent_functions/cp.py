from mythic_payloadtype_container.MythicCommandBase import *
import json


class CpArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {
            "source": CommandParameter(
                name="source",
                type=ParameterType.String,
                description="Source file to copy.",
                ui_position=1
            ),
            "destination": CommandParameter(
                name="destination",
                type=ParameterType.String,
                description="Source will copy to this location",
                ui_position=2
            ),
        }

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("Must provide arguments")
        else:
            try:
                self.load_args_from_json_string(self.command_line)
            except:
                raise Exception("Failed to load arguments as JSON. Did you use the popup?")


class CpCommand(CommandBase):
    cmd = "cp"
    needs_admin = False
    help_cmd = "cp"
    description = "Copy a file from one location to another."
    version = 1
    author = "@xorrior"
    argument_class = CpArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
