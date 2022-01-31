from mythic_payloadtype_container.MythicCommandBase import *
import json


class LibinjectArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
             CommandParameter(
                name="pid",
                display_name="Pid to inject into",
                type=ParameterType.Number,
                description="PID of process to inject into.",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=1
                    )
                ]
            ),
            CommandParameter(
                name="library",
                display_name="Absolute path to dylib on target to load",
                type=ParameterType.String,
                description="Path to the dylib to inject",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=2
                    )
                ]
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class LibinjectCommand(CommandBase):
    cmd = "libinject"
    needs_admin = True
    help_cmd = "libinject"
    description = "Inject a library from on-host into a process."
    version = 1
    author = "@xorrior"
    argument_class = LibinjectArguments
    attackmapping = ["T1055"]
    attributes = CommandAttributes(
        # uncomment when posiedon can dynamically compile commands
        supported_os=[SupportedOS.MacOS]
    )

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        if task.callback.integrity_level <= 2:
            raise Exception("Error: the libinject command requires elevated privileges")
        else:
            return task

    async def process_response(self, response: AgentResponse):
        pass
