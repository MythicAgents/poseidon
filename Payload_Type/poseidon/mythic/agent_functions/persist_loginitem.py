from mythic_payloadtype_container.MythicCommandBase import *
import json

class PersistLoginItemArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="path",
                type=ParameterType.String,
                description="Path to the binary to execute at login",
            ),
            CommandParameter(
                name="name",
                type=ParameterType.String,
                description="The name that is displayed in the Login Items section of the Users & Groups preferences pane",
            ),
            CommandParameter(
                name="global",
                type=ParameterType.Boolean,
                description="Set this to true if the login item should be installed for all users. This requires administrative privileges",
                default_value=True
            ),
        ]

    async def parse_arguments(self):
        if len(self.command_line) > 0:
            if self.command_line[0] == "{":
                self.load_args_from_json_string(self.command_line)
            else:
                raise ValueError("Missing JSON arguments")

        else:
            raise ValueError("Missing arguments")

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class PersistLoginItem(CommandBase):
    cmd = "persist_loginitem"
    needs_admin = False
    help_cmd = "persist_loginitem"
    description = "Add a login item for the current user via the LSSharedFileListInsertItemURL function."
    version = 1
    author = "@xorrior"
    attackmapping = ["T1547.011"]
    argument_class = PersistLoginItemArguments
    attributes = CommandAttributes(
        supported_os=[SupportedOS.MacOS]
    )

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
