from mythic_payloadtype_container.MythicCommandBase import *
import json


class XpcArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="listStatusAction",
                type=ParameterType.ChooseOne,
                description="Pick to list launchd services or the status of a service",
                choices=[
                    "list",
                    "status",
                ],
                parameter_group_info=[
                    ParameterGroupInfo(group_name="list/status"),
                ]
            ),
            CommandParameter(
                name="startStopAction",
                display_name="Pick to start or stop the action",
                type=ParameterType.ChooseOne,
                description="Choose to start or stop the specified service",
                choices=["start", "stop"],
                parameter_group_info=[
                    ParameterGroupInfo(group_name="start/stop")
                ]
            ),
            CommandParameter(
                name="loadUnloadAction",
                display_name="Pick to load or unload the plist",
                description="Path to a property list to either load or unload",
                type=ParameterType.ChooseOne,
                choices=["load", "unload"],
                parameter_group_info=[
                    ParameterGroupInfo(group_name="load/unload")
                ]
            ),
            CommandParameter(
                name="program",
                type=ParameterType.String,
                description="Program/binary to execute if using 'submit' command",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="submit")
                ]
            ),
            CommandParameter(
                name="file",
                display_name="Path to file to load on target",
                type=ParameterType.String,
                description="Path to the plist file if using load/unload commands",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="load/unload")
                ]
            ),
            CommandParameter(
                name="servicename",
                type=ParameterType.String,
                description="Name of the service to communicate with. Used with the submit, send, start/stop commands",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="send"),
                    ParameterGroupInfo(group_name="submit"),
                    ParameterGroupInfo(group_name="start/stop"),
                    ParameterGroupInfo(group_name="list/status"),
                ]
            ),
            CommandParameter(
                name="pid",
                type=ParameterType.Number,
                description="PID of the process",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="procinfo")
                ]
            ),
            CommandParameter(
                name="data",
                type=ParameterType.String,
                description="base64 encoded json data to send to a target service",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="send")
                ]
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class XpcCommand(CommandBase):
    cmd = "xpc"
    needs_admin = False
    help_cmd = "xpc"
    description = "Use xpc to execute routines with launchd or communicate with another service/process."
    version = 1
    author = "@xorrior"
    argument_class = XpcArguments
    attributes = CommandAttributes(
        # uncomment when poseidon can dynamically compile commands
        supported_os=[SupportedOS.MacOS]
    )
    attackmapping = ["T1559"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
